package providers

import "fmt"

type ErrorKind string

const (
	Credentials     ErrorKind = "credentials"
	Disabled        ErrorKind = "disabled"
	Quota           ErrorKind = "quota"
	RateLimit       ErrorKind = "rate_limited"
	Network         ErrorKind = "network"
	Timeout         ErrorKind = "timeout"
	Unavailable     ErrorKind = "unavailable"
	InvalidResponse ErrorKind = "invalid_response"
	InvalidInput    ErrorKind = "invalid_input"
	Service         ErrorKind = "service"
)

// Error carries a sanitized category. Remote payloads and transport details
// deliberately stay out of its representation.
type Error struct {
	Kind      ErrorKind
	Operation string
	Status    int
}

func (e *Error) Error() string {
	if e.Status != 0 {
		return fmt.Sprintf("%s: %s (HTTP %d)", e.Operation, e.Kind, e.Status)
	}
	return fmt.Sprintf("%s: %s", e.Operation, e.Kind)
}
