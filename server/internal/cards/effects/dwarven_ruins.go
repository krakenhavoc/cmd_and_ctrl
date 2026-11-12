package effects

// Dwarven Ruins — Land (Fallen Empires, #5637):
//
//	"This land enters tapped.
//	 {T}: Add {R}.
//	 {T}, Sacrifice this land: Add {R}{R}."
//
// The red one of Fallen Empires' five sacrifice lands, built by
// tappedSacrificeLand (sacrifice_lands.go): a tapped entry, a tap for
// {R}, and a one-shot tap-and-sacrifice for {R}{R}.
//
// No simplification.
func init() {
	Register(tappedSacrificeLand("cbef6ad5-717f-4b02-9c62-e5deba407ad1", "Dwarven Ruins", "R"))
}
