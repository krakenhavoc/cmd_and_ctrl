package effects

// Misty Palms Oasis — Land (Avatar: The Last Airbender, #4459):
//
//	"This land enters tapped.
//	 {T}: Add {W} or {B}.
//	 {4}, {T}, Sacrifice this land: Draw a card."
//
// One of the set's ten common duals, registered through
// avatarB20CyclingTapland (avatar_b20_helpers.go).
//
// No simplification.
func init() {
	Register(avatarB20CyclingTapland("dfd2c57a-4557-4df1-8f6f-da2cbd317f12", "Misty Palms Oasis", "W", "B"))
}
