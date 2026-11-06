package effects

// Airship Engine Room — Land (Avatar: The Last Airbender, #4459):
//
//	"This land enters tapped.
//	 {T}: Add {U} or {R}.
//	 {4}, {T}, Sacrifice this land: Draw a card."
//
// One of the set's ten common duals, registered through
// avatarB20CyclingTapland (avatar_b20_helpers.go).
//
// No simplification.
func init() {
	Register(avatarB20CyclingTapland("01ed9139-293d-4477-a145-3c6fe59266c7", "Airship Engine Room", "U", "R"))
}
