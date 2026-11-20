package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flight Spellbomb — Artifact {1} (Scars of Mirrodin, #5577):
//
//	"{T}, Sacrifice this artifact: Target creature gains flying until end of
//	 turn.
//	 When this artifact is put into a graveyard from the battlefield, you
//	 may pay {U}. If you do, draw a card."
//
// A Spellbomb: the first ability is Air Marshal's grant
// (ebTargetCreatureGainsUntilEOT), the second Nihil Spellbomb's draw
// (somB2SpellbombDraw).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f679093-86a0-4211-8491-de1f6d81c483",
		Name:         "Flight Spellbomb",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice this artifact: Target creature gains flying until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    Plus(TapCost(), SacrificeThis()),
			Targets: TargetCreature("target creature"),
			Effect:  ebTargetCreatureGainsUntilEOT("flying", "Flight Spellbomb — flying until end of turn"),
		}},
		Triggered: []game.TriggeredAbility{somB2SpellbombDraw("Flight Spellbomb", "{U}")},
	})
}
