package flags

import (
	"flag"
	"strings"
	"time"
)

const (
	// KubeconfigFlagDescription = "(optional) absolute path to the kubeconfig file"
	DurationFlagDescription   = "(Optional) Duration between different strategies"
	IncludeAllFlagDescription = "Schedule deletion for all namespaces"
	IncludeFlagDescription    = "Coma-separated list of namespaces to include for deletion"
	IgnoreFlagDescription     = "Coma-separated list of namespaces to ignore from deletion"
)

// ParseFlags parse the flags from the command line
func (flags *FlagsSpec) ParseFlags() {
	// flags.Kubeconfig = flag.String("kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), KubeconfigFlagDescription)

	flags.Duration = flag.Duration("duration-between-strategies", time.Minute, DurationFlagDescription)
	flags.IncludeAll = flag.Bool("include-all", false, IncludeAllFlagDescription)

	includeStr := flag.String("include", "", IncludeFlagDescription)
	flags.Include = strings.Split(strings.TrimSpace(*includeStr), ",")

	ignoreStr := flag.String("ignore", "", IgnoreFlagDescription)
	flags.Ignore = strings.Split(strings.TrimSpace(*ignoreStr), ",")

	flag.Parse()
}
