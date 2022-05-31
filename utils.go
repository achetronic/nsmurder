package main

import "strings"

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
