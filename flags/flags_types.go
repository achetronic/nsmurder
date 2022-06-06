package flags

import (
	"strings"
	"time"
)

// arrayFlags type allows to call a '--flag' more than once
type arrayFlags []string

// String returns a string representation of the type for the 'flag' library
func (i *arrayFlags) String() string {
	result := strings.Join(*i, " ")
	return result
}

// Set defines how an element of the type must be treated when is being set by the 'flag' library
func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// Flags represents all the available command flags
type FlagsSpec struct {
	//Kubeconfig     *string    `json:"kubeconfig"`

	Duration   *time.Duration `json:"duration,omitempty"`
	IncludeAll *bool          `json:"include_all"`
	Include    arrayFlags     `json:"include"`
	Ignore     arrayFlags     `json:"ignore"`
}
