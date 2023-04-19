package springboot

import (
	"fmt"
	"github.com/pkg/errors"
)

type SshError struct {
	error
	message string
}

func (se SshError) Error() string {
	return fmt.Sprintf("failed to execute command, cause: %s, message: %s", se.error, se.message)
}

func (se SshError) Unwrap() error {
	return se.error
}

type PermissionDenied struct {
	error
	message string
}

func (pe PermissionDenied) Error() string {
	return fmt.Sprintf("permission deinied when executing command, cause: %s, message: %s", pe.error, pe.message)
}

func (pe PermissionDenied) Unwrap() error {
	return pe.error
}

type ConnectionError struct {
	error
	message string
}

func (ce ConnectionError) Error() string {
	return fmt.Sprintf("failed to connect to target server, cause: %s, message: %s", ce.error, ce.message)
}

func (ce ConnectionError) Unwrap() error {
	return ce.error
}

type ConnectionTimeoutError struct {
	message string
}

func (ce ConnectionTimeoutError) Error() string {
	return ce.message
}

type CredentialError struct {
	error
	message string
}

func (ce CredentialError) Error() string {
	return fmt.Sprintf("failed to connect with credential, cause: %s, message: %s ", ce.error, ce.message)
}

func (ce CredentialError) Unwrap() error {
	return ce.error
}

func Join(errs ...error) error {
	if len(errs) == 0 {
		return nil
	}

	e := JoinErrors{}
	for _, err := range errs {
		e.errors = append(e.errors, err)
	}
	return e
}

type JoinErrors struct {
	errors []error
}

func (je JoinErrors) Error() string {
	return fmt.Sprintf("joined errors, total %v ", len(je.errors))
}

func (je JoinErrors) Unwrap() []error {
	return je.errors
}

func IsSshError(err error) bool {
	return is(err, &SshError{})
}

func IsConnectionError(err error) bool {
	return is(err, &ConnectionError{})
}

func IsCredentialError(err error) bool {
	return is(err, &CredentialError{})
}

func IsPermissionDenied(err error) bool {
	return is(err, &PermissionDenied{})
}

func IsJoinErrors(err error) bool {
	return is(err, &JoinErrors{})
}

func is(from, to error) bool {
	return errors.As(from, to)
}
