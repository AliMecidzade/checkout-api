package service

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrPaymentDeclined    = errors.New("payment declined")
)
