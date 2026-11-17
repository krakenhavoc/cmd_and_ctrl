package effects

// Havenwood Battleground — Land (Fallen Empires, #5637):
//
//	"This land enters tapped.
//	 {T}: Add {G}.
//	 {T}, Sacrifice this land: Add {G}{G}."
//
// The green one of Fallen Empires' five sacrifice lands, built by
// tappedSacrificeLand (sacrifice_lands.go): a tapped entry, a tap for
// {G}, and a one-shot tap-and-sacrifice for {G}{G}.
//
// No simplification.
func init() {
	Register(tappedSacrificeLand("18cbb47a-85b1-48f6-a024-8c3bbffa0d87", "Havenwood Battleground", "G"))
}
