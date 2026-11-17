package effects

// Ebon Stronghold — Land (Fallen Empires, #5637):
//
//	"This land enters tapped.
//	 {T}: Add {B}.
//	 {T}, Sacrifice this land: Add {B}{B}."
//
// The black one of Fallen Empires' five sacrifice lands, built by
// tappedSacrificeLand (sacrifice_lands.go): a tapped entry, a tap for
// {B}, and a one-shot tap-and-sacrifice for {B}{B}.
//
// No simplification.
func init() {
	Register(tappedSacrificeLand("4d4b9512-d29d-4b99-9249-d37e1199a1f5", "Ebon Stronghold", "B"))
}
