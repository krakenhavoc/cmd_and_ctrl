package effects

// Corpse Traders — Creature — Human Rogue {3}{B}, 3/3:
//
//	"{2}{B}, Sacrifice a creature: Target opponent reveals their hand.
//	 You choose a card from it. That player discards that card.
//	 Activate only as a sorcery."
//
// The revealed-hand pick (ADR 0116) with no filter: any card, a land
// included. The sacrifice is "a creature", so Corpse Traders can pay
// with itself (the 2012-05-01 ruling). The chooser is the activator
// (CR 113.8); "Activate only as a sorcery" is CR 602.5d. A hand that is
// empty is revealed and nothing is discarded (CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e94297af-1236-4c9a-860f-2ee6a0d12e5e",
		Name:         "Corpse Traders",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{2}{B}, Sacrifice a creature: Target opponent reveals their hand. You choose a card from it. That player discards that card. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{2}{B}"), SacrificeACreature()),
			SorcerySpeed: true,
			Targets:      TargetPlayer("target opponent", Opponent()),
			Effect:       TargetRevealsYouChooseDiscardAbility(nil, "card"),
		}},
	})
}
