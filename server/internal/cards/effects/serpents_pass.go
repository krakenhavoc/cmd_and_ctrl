package effects

// Serpent's Pass — Land (Avatar: The Last Airbender, #4459):
//
//	"This land enters tapped.
//	 {T}: Add {U} or {B}.
//	 {4}, {T}, Sacrifice this land: Draw a card."
//
// One of the set's ten common duals, registered through
// avatarB20CyclingTapland (avatar_b20_helpers.go).
//
// No simplification.
func init() {
	Register(avatarB20CyclingTapland("f715f701-a735-42ab-b31a-8e1bd04ac5ff", "Serpent's Pass", "U", "B"))
}
