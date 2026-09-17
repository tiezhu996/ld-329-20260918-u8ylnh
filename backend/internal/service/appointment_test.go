package service

import (
	"sync"
	"testing"

	"cyskillswap/internal/constants"
	"cyskillswap/internal/errors"
	"cyskillswap/internal/model"
	"cyskillswap/internal/repository"
)

func isTerminalConflict(err error) bool {
	biz, ok := err.(errors.BusinessError)
	return ok && (biz.Code == errors.CodeTerminalState || biz.Code == errors.CodeStatusConflict)
}

func newTestService() *AppointmentService {
	return NewAppointmentService(repository.NewAppointmentStore())
}

func createPending(t *testing.T, svc *AppointmentService, initiator, responder string, slots []string) int {
	t.Helper()
	appt, err := svc.Create(CreateAppointmentInput{
		Initiator: initiator, Responder: responder,
		Time: "周六 10:00", Place: "线上会议室", Agenda: "测试", Slots: slots,
	})
	if err != nil {
		t.Fatalf("create appointment: %v", err)
	}
	return appt.ID
}

// holdsSlot 判断用户是否持有指定时间档（种子数据会持有其它档位，故按档位精确判断）。
func holdsSlot(svc *AppointmentService, user, slot string) bool {
	for _, held := range svc.HeldSlots(user) {
		if held.Slot == slot {
			return true
		}
	}
	return false
}

// 待确认时仅发起人可撤回；响应人撤回被拒。
func TestWithdrawOnlyByInitiator(t *testing.T) {
	svc := newTestService()
	id := createPending(t, svc, "林澈", "孟野", []string{"周三晚"})

	if _, err := svc.Withdraw(id, "孟野"); err == nil {
		t.Fatal("responder withdraw should be forbidden")
	}
	if _, err := svc.Withdraw(id, "路人甲"); err == nil {
		t.Fatal("non-participant withdraw should be forbidden")
	}
	res, err := svc.Withdraw(id, "林澈")
	if err != nil {
		t.Fatalf("initiator withdraw: %v", err)
	}
	if !res.Changed || res.Appointment.Status != constants.ApptStatusWithdrawn {
		appt := res.Appointment
		t.Fatalf("withdraw not applied: changed=%v status=%s", res.Changed, appt.Status)
	}
	if res.Appointment.SlotsHeld {
		t.Fatal("slots should be released after withdraw")
	}
}

// 撤回生效后双方时间档立即释放，且同一时间档可重新发起预约。
func TestWithdrawReleasesBothSlots(t *testing.T) {
	svc := newTestService()
	id := createPending(t, svc, "林澈", "孟野", []string{"周三晚"})

	if !holdsSlot(svc, "林澈", "周三晚") {
		t.Fatal("initiator slot should be held")
	}
	if _, err := svc.Withdraw(id, "林澈"); err != nil {
		t.Fatalf("withdraw: %v", err)
	}
	if holdsSlot(svc, "林澈", "周三晚") || holdsSlot(svc, "孟野", "周三晚") {
		t.Fatal("both parties' slots must be released immediately after withdraw")
	}
	// 时间档释放后可立即复用。
	if _, err := svc.Create(CreateAppointmentInput{
		Initiator: "林澈", Responder: "周芮", Time: "t", Place: "p", Slots: []string{"周三晚"},
	}); err != nil {
		t.Fatalf("released slot should be reusable: %v", err)
	}
}

// 重复撤回 / 重复确认不改变终态，刷新后仍可回读。
func TestDuplicateRequestsKeepTerminalState(t *testing.T) {
	svc := newTestService()
	id := createPending(t, svc, "林澈", "孟野", []string{"周三晚"})

	if _, err := svc.Withdraw(id, "林澈"); err != nil {
		t.Fatalf("withdraw: %v", err)
	}
	version := svcMustGet(t, svc, id).Version
	for i := 0; i < 3; i++ {
		res, err := svc.Withdraw(id, "林澈")
		if err != nil {
			t.Fatalf("duplicate withdraw should be idempotent: %v", err)
		}
		if res.Changed {
			t.Fatal("duplicate withdraw must not report changed")
		}
		if _, err := svc.Confirm(id, "孟野"); err == nil {
			t.Fatal("confirm after withdrawn must fail")
		}
	}
	got := svcMustGet(t, svc, id)
	if got.Status != constants.ApptStatusWithdrawn || got.Version != version {
		t.Fatalf("terminal state changed: status=%s version=%d", got.Status, got.Version)
	}
}

// 确认锁定后不能撤回，只能提交取消原因等待对方处理。
func TestConfirmedCannotWithdrawOnlyCancelFlow(t *testing.T) {
	svc := newTestService()
	id := createPending(t, svc, "林澈", "孟野", []string{"周三晚"})

	if _, err := svc.Confirm(id, "林澈"); err != nil {
		t.Fatalf("confirm initiator: %v", err)
	}
	// 单方已确认阶段发起人仍可撤回（待确认窗口内）。
	// 这里继续由响应人确认以锁定。
	if _, err := svc.Confirm(id, "孟野"); err != nil {
		t.Fatalf("confirm responder: %v", err)
	}
	got := svcMustGet(t, svc, id)
	if got.Status != constants.ApptStatusConfirmed {
		t.Fatalf("expected confirmed, got %s", got.Status)
	}
	if _, err := svc.Withdraw(id, "林澈"); err == nil {
		t.Fatal("withdraw after lock must be rejected")
	}
	if _, err := svc.RequestCancel(id, "林澈", "临时有考试冲突"); err != nil {
		t.Fatalf("request cancel: %v", err)
	}
	// 提交取消申请期间时间档仍占用，不能撤回，不能重复确认。
	if !holdsSlot(svc, "孟野", "周三晚") {
		t.Fatal("slots stay held while cancel is pending")
	}
	if _, err := svc.Withdraw(id, "林澈"); err == nil {
		t.Fatal("withdraw while cancel pending must be rejected")
	}
	// 申请人不能自己审批。
	if _, err := svc.DecideCancel(id, "林澈", constants.CancelDecisionApprove); err == nil {
		t.Fatal("self-approval must be rejected")
	}
	// 对方同意后取消生效并释放时间档。
	res, err := svc.DecideCancel(id, "孟野", constants.CancelDecisionApprove)
	if err != nil {
		t.Fatalf("approve cancel: %v", err)
	}
	if res.Appointment.Status != constants.ApptStatusCancelled || res.Appointment.SlotsHeld {
		t.Fatal("cancel should be terminal and release slots")
	}
	if holdsSlot(svc, "林澈", "周三晚") || holdsSlot(svc, "孟野", "周三晚") {
		t.Fatal("slots must be released after approved cancel")
	}
}

// 对方拒绝取消后回到已确认，时间档保持占用。
func TestRejectCancelReturnsToConfirmed(t *testing.T) {
	svc := newTestService()
	id := createPending(t, svc, "林澈", "孟野", []string{"周三晚"})
	_, _ = svc.Confirm(id, "林澈")
	_, _ = svc.Confirm(id, "孟野")
	if _, err := svc.RequestCancel(id, "孟野", "我可能要改期"); err != nil {
		t.Fatalf("request cancel: %v", err)
	}
	res, err := svc.DecideCancel(id, "林澈", constants.CancelDecisionReject)
	if err != nil {
		t.Fatalf("reject cancel: %v", err)
	}
	if res.Appointment.Status != constants.ApptStatusConfirmed {
		t.Fatalf("expected back to confirmed, got %s", res.Appointment.Status)
	}
	if !res.Appointment.SlotsHeld || !holdsSlot(svc, "林澈", "周三晚") {
		t.Fatal("slots must remain held after rejection")
	}
	// 回到 confirmed 后仍不能撤回，只能再次走取消流程。
	if _, err := svc.Withdraw(id, "林澈"); err == nil {
		t.Fatal("withdraw after reject must still be rejected")
	}
}

// 同一方重复确认幂等；双方确认不可被第三次确认改变。
func TestDuplicateConfirmIdempotent(t *testing.T) {
	svc := newTestService()
	id := createPending(t, svc, "林澈", "孟野", []string{"周三晚"})
	r1, err := svc.Confirm(id, "林澈")
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	r2, _ := svc.Confirm(id, "林澈")
	if r2.Changed {
		t.Fatal("same-party duplicate confirm must be idempotent")
	}
	if r1.Appointment.Status != constants.ApptStatusConfirmedPartial {
		t.Fatalf("first confirm should be partial, got %s", r1.Appointment.Status)
	}
	locked, _ := svc.Confirm(id, "孟野")
	if locked.Appointment.Status != constants.ApptStatusConfirmed {
		t.Fatalf("second confirm should lock, got %s", locked.Appointment.Status)
	}
	r3, _ := svc.Confirm(id, "孟野")
	if r3.Changed {
		t.Fatal("confirm after lock must be idempotent")
	}
	if got := svcMustGet(t, svc, id); got.Status != constants.ApptStatusConfirmed {
		t.Fatalf("expected confirmed, got %s", got.Status)
	}
}

// 任一方确认与发起人撤回同时到达时，只能有一个结果。
// 大量并发试次：最终状态必须自洽——撤回则时间档释放，确认则时间档仍占用。
func TestConcurrentConfirmVsWithdrawSingleOutcome(t *testing.T) {
	const trials = 300
	for i := 0; i < trials; i++ {
		svc := newTestService()
		id := createPending(t, svc, "林澈", "孟野", []string{"周三晚"})

		var wg sync.WaitGroup
		wg.Add(2)
		var confirmErr, withdrawErr error
		go func() { defer wg.Done(); _, confirmErr = svc.Confirm(id, "孟野") }()
		go func() { defer wg.Done(); _, withdrawErr = svc.Withdraw(id, "林澈") }()
		wg.Wait()

		got := svcMustGet(t, svc, id)
		switch got.Status {
		case constants.ApptStatusWithdrawn:
			// 撤回是权威终态：
			//  - 撤回先到 -> 并发确认被拒；
			//  - 确认先到（单方已确认）-> 待确认阶段发起人撤回仍生效，覆盖该确认。
			// 两种时序最终都只有 withdrawn 一个结果，且时间档必须释放。
			if got.SlotsHeld || holdsSlot(svc, "孟野", "周三晚") || holdsSlot(svc, "林澈", "周三晚") {
				t.Fatalf("trial %d: withdrawn but slots still held", i)
			}
			if withdrawErr != nil {
				t.Fatalf("trial %d: final withdraw must be the accepted op: %v", i, withdrawErr)
			}
			// 若确认报错，必须是"已撤回"类终态冲突，而不能是其它错误。
			if confirmErr != nil && !isTerminalConflict(confirmErr) {
				t.Fatalf("trial %d: unexpected confirm error: %v", i, confirmErr)
			}
		case constants.ApptStatusConfirmedPartial:
			// 确认胜出意味着撤回必然被拒（撤回只要被状态机接受就一定会成为终态）。
			if withdrawErr == nil {
				t.Fatalf("trial %d: confirm won but withdraw succeeded", i)
			}
			if !got.SlotsHeld || !holdsSlot(svc, "林澈", "周三晚") {
				t.Fatalf("trial %d: confirmed but slots released", i)
			}
		default:
			t.Fatalf("trial %d: unexpected status %s", i, got.Status)
		}
	}
}

// 原有预约保持可用：预置 #1 已确认、#2 待确认，且 #1 不能撤回只能取消。
func TestSeedAppointmentsRemainUsable(t *testing.T) {
	svc := newTestService()
	locked := svcMustGet(t, svc, 1)
	if locked.Status != constants.ApptStatusConfirmed {
		t.Fatalf("seed #1 should be confirmed, got %s", locked.Status)
	}
	if _, err := svc.Withdraw(1, locked.Initiator); err == nil {
		t.Fatal("seed #1 is locked and must not be withdrawable")
	}
	if _, err := svc.RequestCancel(1, locked.Initiator, "原有预约改期协商"); err != nil {
		t.Fatalf("seed #1 should support cancel flow: %v", err)
	}

	pending := svcMustGet(t, svc, 2)
	if pending.Status != constants.ApptStatusPending {
		t.Fatalf("seed #2 should be pending, got %s", pending.Status)
	}
	if !pending.SlotsHeld {
		t.Fatal("seed #2 slots should be held")
	}
}

func svcMustGet(t *testing.T, svc *AppointmentService, id int) model.Appointment {
	t.Helper()
	appt, err := svc.Get(id)
	if err != nil {
		t.Fatalf("get appointment %d: %v", id, err)
	}
	return appt
}
