package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Honor-Worn Shaku — Artifact {3}:
//
//	"{T}: Add {C}.
//	 Tap an untapped legendary permanent you control: Untap this
//	 artifact."
//
// The untap's cost is "tap an untapped legendary permanent you control"
// (TapAnUntapped), not the {T} symbol (CR 302.6), so a legendary creature
// that arrived this turn may pay it, and it does not target. The Shaku is
// not legendary, so it can never pay for itself. The untap is an
// ordinary activated ability that uses the stack (it adds no mana, CR
// 605.1a), Basalt Monolith's shape.
//
// No simplification. Guffster's Party deck request (#4184).
func init() {
	Register(Spec{
		OracleID:     "babeeaf6-0fe4-491f-bbf5-63a0568b3d6e",
		Name:         "Honor-Worn Shaku",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			UntapSelf: true,
			Label:     "Tap an untapped legendary permanent you control: Untap this artifact.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      TapAnUntapped("an untapped legendary permanent you control", Legendary()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return UntapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
