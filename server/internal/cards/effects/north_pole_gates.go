package effects

// North Pole Gates — Land (Avatar: The Last Airbender, #4459):
//
//	"This land enters tapped.
//	 {T}: Add {W} or {U}.
//	 {4}, {T}, Sacrifice this land: Draw a card."
//
// One of the set's ten common duals, registered through
// avatarB20CyclingTapland (avatar_b20_helpers.go).
//
// No simplification.
func init() {
	Register(avatarB20CyclingTapland("3fcd1432-66d0-40e1-85ef-b57a6fa94b7c", "North Pole Gates", "W", "U"))
}
