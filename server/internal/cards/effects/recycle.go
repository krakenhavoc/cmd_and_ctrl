package effects

// Recycle — Enchantment {4}{G}{G}:
//
//	"Skip your draw step.
//	 Whenever you play a card, draw a card.
//	 Your maximum hand size is two."
//
// Null Profusion's green twin, built by the same playACardHandOfTwo
// (max_hand_size.go). "Play a card" is a land played or a spell cast
// that is a card, never a copy (the 2018-07-13 ruling); a countered
// spell still drew. The maximum is a maximum-hand-size static folded in
// CR 613.11 timestamp order (ADR 0113 §3, #2074).
//
// No simplification.
func init() {
	Register(playACardHandOfTwo("8cef3ce7-8fbc-4e68-93f9-d6adfc2c2cbf", "Recycle"))
}
