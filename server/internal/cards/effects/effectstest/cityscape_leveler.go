package effectstest

func HasKeywordForTest(abilities []string, kw string) bool {
	for _, a := range abilities {
		if a == kw {
			return true
		}
	}
	return false
}
