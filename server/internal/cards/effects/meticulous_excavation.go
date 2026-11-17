package effects

// Meticulous Excavation — Enchantment {W} (The Brothers' War, #5482):
//
//	"{2}{W}: Return target permanent you control to its owner's hand. If it
//	 has unearth, instead exile it, then return that card to its owner's
//	 hand. Activate only during your turn."
//
// The target is chosen on activation and re-checked as the ability resolves
// (CR 608.2b). Whether the permanent has unearth is read then, off the
// permanent as it is (s82Bro13HasUnearth). With unearth it is exiled and the
// card that reached exile returns to its owner's hand, so it is a card again
// and may be unearthed once more from a graveyard later; a token is exiled and
// ceases to exist. A permanent without unearth is simply bounced.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6dff1d15-604f-4eea-9b7a-0c0ba6afbbac",
		Name:         "Meticulous Excavation",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{2}{W}: Return target permanent you control to its owner's hand. If it has unearth, instead exile it, then return that card to its owner's hand. Activate only during your turn.",
			Cost:      ManaCost("{2}{W}"),
			Targets:   TargetPermanent("target permanent you control", YouControl()),
			Condition: DuringYourTurn(),
			Effect:    s82Bro13ReturnOrExileToHand,
		}},
	})
}
