package response

type HTTPError struct {
	Status int
	Msg    string
}

func NewError(status int, msg string) HTTPError {
	return HTTPError{Status: status, Msg: msg}
}

func (e HTTPError) Error() string {
	return e.Msg
}

func (e HTTPError) HTTPStatus() int {
	return e.Status
}
