package effects

// Reckless Scholar — Creature — Human Wizard {2}{U}, 2/1 (Foundations
// Jumpstart, S82 BNS):
//
//	"{T}: Target player draws a card, then discards a card."
//
// Cephalid Looter's ability through the shared odyTargetPlayerLoots: any
// player may be the target, the draw comes first, and the discard is chosen
// from the hand it made.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "604aded2-64c2-4474-b5b7-8c83e0e3fe76",
		Name:         "Reckless Scholar",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{odyTargetPlayerLoots("{T}: Target player draws a card, then discards a card.", 1)},
	})
}
