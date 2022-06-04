package flags

// StringInList check existance of a string into a list of strings
func StringInList(item string, list []string) (found bool) {
	for _, listItem := range list {
		if listItem == item {
			return true
		}
	}

	return false
}
