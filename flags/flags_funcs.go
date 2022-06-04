package flags

import (
	"flag"
)

const (
	// KubeconfigFlagDescription = "(optional) absolute path to the kubeconfig file"
	IncludeAllFlagDescription = "Schedule deletion for all namespaces"
	IncludeFlagDescription    = "Namespaces to include in deletion list"
	IgnoreFlagDescription     = "Namespaces to ignore from deletion list"
)

// ParseFlags parse the flags from the command line
func (flags *FlagsSpec) ParseFlags() {
	// flags.Kubeconfig = flag.String("kubeconfigg", filepath.Join(homedir.HomeDir(), ".kube", "config"), KubeconfigFlagDescription)
	flags.IncludeAll = flag.Bool("include-all", false, IncludeAllFlagDescription)
	flag.Var(&flags.Include, "include", IncludeFlagDescription)
	flag.Var(&flags.Ignore, "ignore", IgnoreFlagDescription)
	flag.Parse()
}

// GetNamespacesFromFlags return a list with namespaces already filtered from flags
func (flags *FlagsSpec) GetNamespaces() (namespaces []string) {

	for _, includedNamespace := range flags.Include {

		// Ignore desired namespaces
		if StringInList(includedNamespace, flags.Ignore) {
			continue
		}

		namespaces = append(namespaces, includedNamespace)
	}

	return namespaces
}
