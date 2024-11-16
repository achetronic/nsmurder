package flags

import (
	"time"
)

// FlagsSpec represents all the available command flags
type FlagsSpec struct {
	Duration   *time.Duration
	IncludeAll *bool
	Include    []string
	Ignore     []string
}
