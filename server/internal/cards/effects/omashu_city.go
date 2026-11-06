package effects

// Omashu City — Land (Avatar: The Last Airbender, #4459):
//
//	"This land enters tapped.
//	 {T}: Add {R} or {G}.
//	 {4}, {T}, Sacrifice this land: Draw a card."
//
// One of the set's ten common duals, registered through
// avatarB20CyclingTapland (avatar_b20_helpers.go).
//
// No simplification.
func init() {
	Register(avatarB20CyclingTapland("6f9a4bb2-5959-4a5c-83af-c2729e6673df", "Omashu City", "R", "G"))
}
