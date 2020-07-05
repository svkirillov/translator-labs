package helpers

func Unique(slice []string) []string {
	uniqMap := make(map[string]struct{})

	for i := range slice {
		uniqMap[slice[i]] = struct{}{}
	}

	uniqSlice := make([]string, 0)

	for k, _ := range uniqMap {
		uniqSlice = append(uniqSlice, k)
	}

	return uniqSlice
}
