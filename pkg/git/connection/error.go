package connection

import (
	"fmt"
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
)

// ValidationError holds message and reason of an error that occurred during a connection validation
type ValidationError interface {
	error
	Reason() v1alpha1.ConnectionFailureReason
}

type validationError struct {
	message string
	reason  v1alpha1.ConnectionFailureReason
}

func (e *validationError) Error() string {
	return e.message
}

func (e *validationError) Reason() v1alpha1.ConnectionFailureReason {
	return e.reason
}

func newValidationErrorf(reason v1alpha1.ConnectionFailureReason, message string, args ...interface{}) ValidationError {
	return &validationError{message: fmt.Sprintf(message, args...), reason: reason}
}
