package service

import (
	"path/filepath"
	"sync"
	"testing"

	"cyskillswap/internal/constants"
	"cyskillswap/internal/errors"
	"cyskillswap/internal/repository"
)

// newTestStore 为每个用例准备独立的文件存储，避免互相污染。
func newTestStore(t *testing.T) (*repository.AppointmentStore, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "appointments.json")
	store := repository.NewAppointmentStore(path)
	repository.SetAppointmentsForTest(store)
	t.Cleanup(func() { repository.SetAppointmentsForTest(nil) })
	return store, path
}

func mustBusinessError(t *testing.T, err error, code string) {
	t.Helper()
	biz, ok := err.(errors.BusinessError)
	if !ok {
		t.Fatalf("expected BusinessError, got %v", err)
	}
	if biz.Code != code {
		t.Fatalf("expected code %s, got %s (%s)", code, biz.Code, biz.Message)
	}
}

// 原有预约保持可用：种子数据可回读，状态正确映射到新状态机。
func TestSeedAppointmentsRemainUsable(t *testing.T) {
	newTestStore(t)

	confirmed, err := GetAppointment(1)
	if err != nil {
		t.Fatalf("get appointment 1: %v", err)
	}
	if confirmed.State != constants.AppointmentStateConfirmed || confirmed.Status != constants.AppointmentLabelConfirmed {
		t.Fatalf("appointment 1 state mismatch: %+v", confirmed)
	}

	pending, err := GetAppointment(2)
	if err != nil {
		t.Fatalf("get appointment 2: %v", err)
	}
	if pending.State != constants.AppointmentStatePending || pending.Status != constants.AppointmentLabelPending {
		t.Fatalf("appointment 2 state mismatch: %+v", pending)
	}

	if list := ListAppointments(); len(list) != 4 {
		t.Fatalf("expected 4 seeded appointments, got %d", len(list))
	}
}

// 待确认时仅发起人可撤回：对方撤回被拒绝，发起人撤回成功。
func TestWithdrawOnlyInitiator(t *testing.T) {
	newTestStore(t)

	// ID 4 发起人孟野，对方林澈撤回 → 403
	if _, err := WithdrawAppointment(4, "林澈"); err == nil {
		t.Fatal("participant withdraw should fail")
	} else {
		mustBusinessError(t, err, constants.ErrCodeOnlyInitiatorWithdraw)
	}

	// 无关用户 → 403 NOT_PARTY
	if _, err := WithdrawAppointment(4, "许安"); err == nil {
		t.Fatal("outsider withdraw should fail")
	} else {
		mustBusinessError(t, err, constants.ErrCodeNotParty)
	}

	// 发起人撤回 → 成功
	appt, err := WithdrawAppointment(4, "孟野")
	if err != nil {
		t.Fatalf("initiator withdraw: %v", err)
	}
	if appt.State != constants.AppointmentStateWithdrawn || !appt.SlotReleased {
		t.Fatalf("expected withdrawn with released slot, got %+v", appt)
	}
}

// 撤回生效后双方时间档立即释放。
func TestWithdrawReleasesBothSlots(t *testing.T) {
	newTestStore(t)

	for _, user := range []string{"林澈", "周芮"} {
		if len(SlotLocks(user)) == 0 {
			t.Fatalf("expected locked slots for %s before withdraw", user)
		}
	}

	if _, err := WithdrawAppointment(3, "林澈"); err != nil {
		t.Fatalf("withdraw: %v", err)
	}

	for _, user := range []string{"林澈", "周芮"} {
		for _, lock := range SlotLocks(user) {
			if lock.AppointmentID == 3 {
				t.Fatalf("slot of appointment 3 should be released for %s", user)
			}
		}
	}
}

// 确认锁定后不能撤回，只能提交取消原因等待对方处理。
func TestConfirmedAppointmentMustUseCancelFlow(t *testing.T) {
	newTestStore(t)

	// ID 1 已双方确认：发起人撤回被拒绝
	if _, err := WithdrawAppointment(1, "林澈"); err == nil {
		t.Fatal("withdraw on confirmed appointment should fail")
	} else {
		mustBusinessError(t, err, constants.ErrCodeLocked)
	}

	// 取消原因必填
	if _, err := RequestCancel(1, "林澈", "  "); err == nil {
		t.Fatal("empty reason should fail")
	} else {
		mustBusinessError(t, err, constants.ErrCodeCancelReasonRequired)
	}

	// 提交取消原因 → 等待对方处理
	appt, err := RequestCancel(1, "林澈", "临时有课，需要改期")
	if err != nil {
		t.Fatalf("request cancel: %v", err)
	}
	if appt.State != constants.AppointmentStateCancelPending || appt.CancelRequest == nil {
		t.Fatalf("expected cancel_pending, got %+v", appt)
	}

	// 取消处理中不能确认、不能撤回
	if _, err := ConfirmAppointment(1, "孟野"); err == nil {
		t.Fatal("confirm during cancel_pending should fail")
	} else {
		mustBusinessError(t, err, constants.ErrCodeCancelInProgress)
	}
	if _, err := WithdrawAppointment(1, "林澈"); err == nil {
		t.Fatal("withdraw during cancel_pending should fail")
	} else {
		mustBusinessError(t, err, constants.ErrCodeCancelInProgress)
	}

	// 申请人自己不能处理 → 对方拒绝 → 回到已确认
	if _, err := RespondCancel(1, "林澈", constants.CancelActionAccept); err == nil {
		t.Fatal("requester responding own cancel should fail")
	} else {
		mustBusinessError(t, err, constants.ErrCodeNotCancelResponder)
	}
	rejected, err := RespondCancel(1, "孟野", constants.CancelActionReject)
	if err != nil {
		t.Fatalf("reject cancel: %v", err)
	}
	if rejected.State != constants.AppointmentStateConfirmed || rejected.CancelRequest.Status != constants.CancelStatusRejected {
		t.Fatalf("expected confirmed after reject, got %+v", rejected)
	}

	// 再次提交取消 → 对方同意 → 已取消并释放时间档
	if _, err := RequestCancel(1, "林澈", "还是改到下学期"); err != nil {
		t.Fatalf("request cancel again: %v", err)
	}
	cancelled, err := RespondCancel(1, "孟野", constants.CancelActionAccept)
	if err != nil {
		t.Fatalf("accept cancel: %v", err)
	}
	if cancelled.State != constants.AppointmentStateCancelled || !cancelled.SlotReleased {
		t.Fatalf("expected cancelled with released slot, got %+v", cancelled)
	}
	for _, user := range []string{"林澈", "孟野"} {
		for _, lock := range SlotLocks(user) {
			if lock.AppointmentID == 1 {
				t.Fatalf("slot of appointment 1 should be released for %s", user)
			}
		}
	}
}

// 双方确认流程：对方确认后状态锁定为已确认。
func TestConfirmLocksAppointment(t *testing.T) {
	newTestStore(t)

	appt, err := ConfirmAppointment(4, "林澈")
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if appt.State != constants.AppointmentStateConfirmed || !appt.BothConfirmed() {
		t.Fatalf("expected confirmed, got %+v", appt)
	}
}

// 重复请求不能改变终态：终态下同操作幂等返回，冲突操作失败且状态不变。
func TestTerminalStateIsStable(t *testing.T) {
	newTestStore(t)

	first, err := WithdrawAppointment(3, "林澈")
	if err != nil {
		t.Fatalf("withdraw: %v", err)
	}

	// 重复撤回：幂等返回，版本不变
	again, err := WithdrawAppointment(3, "林澈")
	if err != nil {
		t.Fatalf("repeat withdraw should be idempotent: %v", err)
	}
	if again.State != constants.AppointmentStateWithdrawn || again.Version != first.Version {
		t.Fatalf("terminal state changed by repeat withdraw: %+v", again)
	}

	// 终态上的冲突操作全部失败，且状态不变
	conflictOps := []func() error{
		func() error { _, err := ConfirmAppointment(3, "周芮"); return err },
		func() error { _, err := RequestCancel(3, "林澈", "原因"); return err },
		func() error { _, err := RespondCancel(3, "周芮", constants.CancelActionAccept); return err },
	}
	for i, op := range conflictOps {
		if err := op(); err == nil {
			t.Fatalf("conflict op %d on terminal state should fail", i)
		}
	}
	final, err := GetAppointment(3)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if final.State != constants.AppointmentStateWithdrawn || final.Version != first.Version {
		t.Fatalf("terminal state mutated: %+v", final)
	}
}

// 任一方确认与发起人撤回同时到达时只能一个结果。
func TestConcurrentConfirmAndWithdrawOnlyOneWins(t *testing.T) {
	for round := 0; round < 20; round++ {
		newTestStore(t)

		const workers = 8
		var wg sync.WaitGroup
		confirmOK := make(chan bool, workers)
		withdrawOK := make(chan bool, workers)

		for i := 0; i < workers; i++ {
			wg.Add(2)
			go func() {
				defer wg.Done()
				_, err := ConfirmAppointment(4, "林澈") // 对方确认即锁定
				confirmOK <- err == nil
			}()
			go func() {
				defer wg.Done()
				_, err := WithdrawAppointment(4, "孟野") // 发起人撤回
				withdrawOK <- err == nil
			}()
		}
		wg.Wait()
		close(confirmOK)
		close(withdrawOK)

		confirmed, withdrawn := 0, 0
		for ok := range confirmOK {
			if ok {
				confirmed++
			}
		}
		for ok := range withdrawOK {
			if ok {
				withdrawn++
			}
		}

		final, err := GetAppointment(4)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		switch final.State {
		case constants.AppointmentStateConfirmed:
			// 确认先生效：撤回必须全部失败，时间档不释放
			if withdrawn != 0 {
				t.Fatalf("round %d: withdraw succeeded after confirm locked", round)
			}
			if final.SlotReleased {
				t.Fatalf("round %d: confirmed appointment must not release slot", round)
			}
		case constants.AppointmentStateWithdrawn:
			// 撤回先生效：确认必须全部失败，时间档已释放
			if confirmed != 0 {
				t.Fatalf("round %d: confirm succeeded after withdraw", round)
			}
			if !final.SlotReleased {
				t.Fatalf("round %d: withdrawn appointment must release slot", round)
			}
		default:
			t.Fatalf("round %d: unexpected final state %s", round, final.State)
		}
	}
}

// 刷新后仍可回读：状态变更落盘，新存储实例（模拟重启/刷新）能读到最新状态。
func TestStatePersistsAcrossReload(t *testing.T) {
	_, path := newTestStore(t)

	if _, err := WithdrawAppointment(3, "林澈"); err != nil {
		t.Fatalf("withdraw: %v", err)
	}
	if _, err := RequestCancel(1, "林澈", "临时有课"); err != nil {
		t.Fatalf("request cancel: %v", err)
	}

	// 用同一路径重建存储，模拟服务重启后的回读
	reloaded := repository.NewAppointmentStore(path)
	repository.SetAppointmentsForTest(reloaded)

	appt3, err := GetAppointment(3)
	if err != nil {
		t.Fatalf("get 3: %v", err)
	}
	if appt3.State != constants.AppointmentStateWithdrawn || !appt3.SlotReleased {
		t.Fatalf("appointment 3 not persisted: %+v", appt3)
	}
	appt1, err := GetAppointment(1)
	if err != nil {
		t.Fatalf("get 1: %v", err)
	}
	if appt1.State != constants.AppointmentStateCancelPending || appt1.CancelRequest.Reason != "临时有课" {
		t.Fatalf("appointment 1 not persisted: %+v", appt1)
	}
}
