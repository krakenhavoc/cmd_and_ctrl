package effects

// Pilfering Imp — Creature — Imp {B}, 1/1:
//
//	"Flying
//	 {1}{B}, {T}, Sacrifice this creature: Target opponent reveals their
//	 hand. You choose a nonland card from it. That player discards that
//	 card. Activate only as a sorcery."
//
// Thoughtseize's pick (ADR 0116) on an activated ability whose cost
// sacrifices its source. The chooser is the ability's controller, the
// player who activated it (CR 113.8), so the Imp being in the graveyard
// by resolution changes nothing. "Activate only as a sorcery" is the
// sorcery timing of CR 602.5d. The {T} needs the Imp free of summoning
// sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "61017a67-653c-4f75-9ec3-22d2149340a8",
		Name:            "Pilfering Imp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:        "{1}{B}, {T}, Sacrifice this creature: Target opponent reveals their hand. You choose a nonland card from it. That player discards that card. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{1}{B}"), TapCost(), SacrificeThis()),
			SorcerySpeed: true,
			Targets:      TargetPlayer("target opponent", Opponent()),
			Effect:       TargetRevealsYouChooseDiscardAbility(Nonland(), "nonland card"),
		}},
	})
}
