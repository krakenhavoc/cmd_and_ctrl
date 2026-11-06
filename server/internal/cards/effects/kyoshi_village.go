package effects

// Kyoshi Village — Land (Avatar: The Last Airbender, #4459):
//
//	"This land enters tapped.
//	 {T}: Add {G} or {W}.
//	 {4}, {T}, Sacrifice this land: Draw a card."
//
// One of the set's ten common duals, registered through
// avatarB20CyclingTapland (avatar_b20_helpers.go).
//
// No simplification.
func init() {
	Register(avatarB20CyclingTapland("da81b78b-89fd-4fc1-9395-b5b01b96be70", "Kyoshi Village", "G", "W"))
}
