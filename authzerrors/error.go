package authzerrors

import "errors"

var ErrDeniedByPolicy = errors.New("denied by policy agent")
