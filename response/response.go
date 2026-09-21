package response

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

type Response[T any] struct {
	status int

	Msg string `json:"msg"`
	Data T    `json:"data,omitempty"`
}

func NewResponse[T any](msg string, data T) Response[T] {
	return Response[T]{
		status: http.StatusOK,
		Msg: msg,
		Data: data,
	}
}

func (r Response[T]) WithStatus(status int) Response[T] {
	r.status = status
	return r
}

func Error(status int, msg string) Response[gin.H] {
	return NewResponse[gin.H](msg, nil).WithStatus(status)
}
func (r Response[T]) Error() string {
	return r.Msg
}

func (r Response[T]) Status() int {
	return r.status
}

func (r Response[T]) Write(c *gin.Context) error {
	switch {
	case r.status == 0:
		r.status = http.StatusOK
	case r.status == http.StatusUnauthorized:
		c.Header("WWW-Authenticate", "Bearer")
	case r.status >= 400:
		c.Header("Cache-Control", "no-store")
	}

	c.JSON(r.status, r)
	return nil
}

func Handle(c *gin.Context, err error) {
	var resp interface {
		error
		Status() int
	}
	status, msg := http.StatusInternalServerError, "Unhandled request error"
	if errors.As(err, &resp) {
		status, msg = resp.Status(), resp.Error()
	}
	_ = NewResponse[gin.H](msg, gin.H{"traceId": requestid.Get(c)}).WithStatus(status).Write(c)
}



func Decode(c *gin.Context, target any) error {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
	d := json.NewDecoder(c.Request.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return Error(http.StatusBadRequest, "Invalid JSON request")
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return Error(http.StatusBadRequest, "Expected a single JSON object")
	}
	return nil
}
