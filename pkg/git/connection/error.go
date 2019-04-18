package connection

import (
	"fmt"
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
)

type ValidationError struct {
	Message string
	Reason  v1alpha1.ConnectionFailureReason
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("message: %s, reason: %s", e.Message, e.Reason)
}

func newValidationErrorf(reason v1alpha1.ConnectionFailureReason, message string, args ...interface{}) *ValidationError {
	return &ValidationError{Message: fmt.Sprintf(message, args...), Reason: reason}
}
