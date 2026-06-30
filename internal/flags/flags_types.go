// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

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
