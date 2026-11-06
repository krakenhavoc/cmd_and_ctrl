package effects

// Foggy Bottom Swamp — Land (Avatar: The Last Airbender, #4459):
//
//	"This land enters tapped.
//	 {T}: Add {B} or {G}.
//	 {4}, {T}, Sacrifice this land: Draw a card."
//
// One of the set's ten common duals, registered through
// avatarB20CyclingTapland (avatar_b20_helpers.go).
//
// No simplification.
func init() {
	Register(avatarB20CyclingTapland("b56bb8cb-457e-4b85-9bf9-5e891e79f1dd", "Foggy Bottom Swamp", "B", "G"))
}
