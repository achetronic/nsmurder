package flags

import (
	"flag"
	"strings"
	"time"
)

const (
	DurationFlagDescription   = "(Optional) Duration between different strategies"
	IncludeAllFlagDescription = "Schedule deletion for all namespaces."
	IncludeFlagDescription    = "Coma-separated list of namespaces to include for deletion"
	IgnoreFlagDescription     = "Coma-separated list of namespaces to ignore from deletion"
)

// ParseFlags parse the flags from the command line
func (flags *FlagsSpec) ParseFlags() {

	flags.Duration = flag.Duration("duration-between-strategies", time.Minute, DurationFlagDescription)
	flags.IncludeAll = flag.Bool("include-all", false, IncludeAllFlagDescription)

	includeStr := flag.String("include", "", IncludeFlagDescription)
	ignoreStr := flag.String("ignore", "", IgnoreFlagDescription)

	flag.Parse()

	if *includeStr != "" {
		flags.Include = strings.Split(strings.TrimSpace(*includeStr), ",")
	}

	if *ignoreStr != "" {
		flags.Ignore = strings.Split(strings.TrimSpace(*ignoreStr), ",")
	}
}
