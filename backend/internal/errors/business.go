package errors

type BusinessError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e BusinessError) Error() string { return e.Message }
