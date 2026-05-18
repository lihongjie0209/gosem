// Copyright (c) 2022 Circutor S.A. All rights reserved.

package dlms

type ErrorCode int

const (
	ErrorUnspecified ErrorCode = iota
	ErrorInvalidState
	ErrorInvalidParameter
	ErrorCommunicationFailed
	ErrorInvalidResponse
	ErrorAuthenticationFailed
	ErrorInvalidPassword
	ErrorWrongKeys
	ErrorFailureInvocationCounter
	ErrorGetRejected
	ErrorSetRejected
	ErrorActionRejected
	ErrorSetPartial
	ErrorCheckDoesNotMatch
	ErrorPartialTransfer
)

type Error struct {
	code ErrorCode
	msg  string
}

func NewError(code ErrorCode, msg string) *Error {
	return &Error{
		code: code,
		msg:  msg,
	}
}

func (ce *Error) Error() string {
	return ce.msg
}

func (ce *Error) Code() ErrorCode {
	return ce.code
}
