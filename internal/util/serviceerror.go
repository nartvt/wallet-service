package util

import "github.com/go-kratos/kratos/v2/errors"

func BadRequestError(err error) *errors.Error {
	return errors.BadRequest("", "")
}

func UnAuthorizeError() *errors.Error {
	return errors.Unauthorized("", "")
}

func InternalServerError(err error) *errors.Error {
	return errors.InternalServer(err.Error(), "")
}
