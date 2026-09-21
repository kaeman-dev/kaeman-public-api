package response

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response[T any] struct {
	status int

	Msg  string `json:"msg"`
	Data T      `json:"data,omitempty"`
}

func New(msg string) Response[any] {
	return Response[any]{
		Msg: msg,
	}
}

func NewData[T any](msg string, data T) Response[T] {
	return Response[T]{
		Msg:  msg,
		Data: data,
	}
}

func (r Response[T]) WithStatus(status int) Response[T] {
	r.status = status
	return r
}

func (r Response[T]) Error() string {
	return r.Msg
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

func Decode[T any](c *gin.Context, target *T) error {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
	d := json.NewDecoder(c.Request.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		if err == nil {
			return errors.New("body must contain single JSON value")
		}
		return err
	}
	return nil
}
