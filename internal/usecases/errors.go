package usecases

import "errors"

var ErrInvalidCredentials = errors.New("credentials are invalid")

var ErrCustomerNotFound = errors.New("customer not found")
