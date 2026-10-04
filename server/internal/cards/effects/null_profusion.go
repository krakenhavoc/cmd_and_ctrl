package effects

// Null Profusion — Enchantment {4}{B}{B}:
//
//	"Skip your draw step.
//	 Whenever you play a card, draw a card.
//	 Your maximum hand size is two."
//
// Built by playACardHandOfTwo (max_hand_size.go), which Recycle shares.
// "Play a card" is a land played or a spell cast, never a copy (the
// 2007-02-01 ruling; YouPlayedACard). A countered spell still drew.
// "Your maximum hand size is two" is a maximum-hand-size static (ADR
// 0113 §3, #2074), folded in CR 613.11's timestamp order: a Spellbook
// that was there first leaves you at two, one that entered later lifts
// the maximum (the 2009-10-01 ruling).
//
// No simplification.
func init() {
	Register(playACardHandOfTwo("7077ee9a-59c6-4b8e-bf33-9b094412ad38", "Null Profusion"))
}
