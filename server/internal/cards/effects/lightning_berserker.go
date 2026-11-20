package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lightning Berserker — Creature — Human Berserker {R}, 1/1 (Dragons of Tarkir,
// #5551):
//
//	"{R}: This creature gets +1/+0 until end of turn.
//	 Dash {R} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)"
//
// Dash is the shared alternative cost (Ragavan's, CR 702.109): cast for
// its dash cost it gains haste and returns to its owner's hand at the
// beginning of the next end step.
// The firebreathing is Chartooth Cougar's ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "1e474772-3fe4-46be-8c5f-6e477b9fd74e",
		Name:             "Lightning Berserker",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Dash("{R}")},
		Activated: []ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{R}"),
			Effect:  thisGetsUntilEndOfTurn(1, 0, "Lightning Berserker — +1/+0 until end of turn"),
		}},
	})
}
