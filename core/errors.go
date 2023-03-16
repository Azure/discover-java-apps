package core

import (
	"fmt"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"microsoft.com/azure-spring-discovery/api/v1alpha1"
)

type RetryableError struct {
	error
}

func (r RetryableError) Error() string {
	return r.error.Error()
}

type SshError struct {
	error
	message string
}

func (se SshError) Error() string {
	return fmt.Sprintf("run command got error %s, output: %s", se.error.Error(), se.message)
}

type PermissionDenied struct {
	error
	message string
}

func (pe PermissionDenied) Error() string {
	if pe.error != nil {
		return fmt.Sprintf("Run command failed %s, output: %s", pe.error.Error(), pe.message)
	}
	return pe.message
}

type ConnectionError struct {
	error
	message string
}

func (ce ConnectionError) Error() string {
	if ce.error != nil {
		return fmt.Sprintf("%s, error: %s", ce.message, ce.error.Error())
	}

	return fmt.Sprintf("cannot connect to target server, error: %s", ce.message)
}

type OverMaxTimeError struct {
	message string
}

func (oe OverMaxTimeError) Error() string {
	return oe.message
}

type CredentialError struct {
	error
	message string
}

func (ce CredentialError) Error() string {
	return ce.message
}

func Join(errs ...error) error {
	if len(errs) == 0 {
		return nil
	}

	e := &joinErrors{}
	for _, err := range errs {
		e.errors = append(e.errors, err)
	}
	return e
}

type joinErrors struct {
	errors []error
}

func (es joinErrors) Error() string {
	return fmt.Sprintf("contains error count: " + strconv.Itoa(len(es.errors)))
}

type DiscoveryError struct {
	error
	message  string
	severity v1alpha1.Severity
}

func (de DiscoveryError) Error() string {
	if de.error != nil {
		return fmt.Sprintf("%s, error: %s", de.message, de.error.Error())
	}

	return fmt.Sprintf("error occurred during discovery, error: %s", de.message)
}

func mapErrors(runAsAccountId string, errs ...error) []v1alpha1.ErrorObj {
	var errors []v1alpha1.ErrorObj
	if errs == nil {
		return errors
	}

	for _, err := range errs {
		switch err.(type) {
		case joinErrors:
			errors = append(errors, mapErrors(runAsAccountId, err.(joinErrors).errors...)...)
		default:
			errors = append(errors, mapError(err, runAsAccountId))
		}
	}

	return errors
}

func mapError(err error, runAsAccountId string) v1alpha1.ErrorObj {
	var errObj v1alpha1.ErrorObj
	switch err.(type) {
	case ConnectionError:
		errObj = v1alpha1.ConnectionError
	case PermissionDenied:
		errObj = v1alpha1.PermissionError
	case SshError:
		errObj = v1alpha1.SshCmdError
	case CredentialError:
		errObj = v1alpha1.CredentialError
	case DiscoveryError:
		de := err.(DiscoveryError)
		errorObj := v1alpha1.GeneralError
		errorObj.PossibleCause = de.message
		errorObj.Severity = de.severity
		errObj = errorObj
	case joinErrors:
		errObj = v1alpha1.MultipleError
	default:
		errObj = v1alpha1.GeneralError
	}

	errObj.Message = err.Error()
	errObj.UpdatedTimeStamp = metav1.NewTime(time.Now())
	errObj.RunAsAccountId = runAsAccountId
	return errObj
}
