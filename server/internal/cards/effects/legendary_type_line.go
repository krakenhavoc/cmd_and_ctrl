package effects

import "strings"

// LegendaryTypeLine adds the Legendary supertype to a type line that
// does not already carry it — "except it's legendary". Supertypes
// lead the line, so it is prepended.
func LegendaryTypeLine(tl string) string {
	for _, w := range strings.Fields(tl) {
		if w == "Legendary" {
			return tl
		}
	}
	if tl == "" {
		return "Legendary"
	}
	return "Legendary " + tl
}
