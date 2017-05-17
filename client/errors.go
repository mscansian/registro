package client

import (
	"errors"
	"strconv"
)

var (
	ErrAppNotExist  = errors.New("application doesn't exist.")
	ErrInstNotExist = errors.New("instance doesn't exist.")
)

type UnexpectedCodeError struct {
	Code int
}

func (e *UnexpectedCodeError) Error() string { return "unexpected http code: " + strconv.Itoa(e.Code) }
