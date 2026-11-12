package effects

// Svyelunite Temple — Land (Fallen Empires, #5637):
//
//	"This land enters tapped.
//	 {T}: Add {U}.
//	 {T}, Sacrifice this land: Add {U}{U}."
//
// The blue one of Fallen Empires' five sacrifice lands, built by
// tappedSacrificeLand (sacrifice_lands.go): a tapped entry, a tap for
// {U}, and a one-shot tap-and-sacrifice for {U}{U}.
//
// No simplification.
func init() {
	Register(tappedSacrificeLand("5578cf33-62f5-456c-a58c-f744a25df79b", "Svyelunite Temple", "U"))
}
