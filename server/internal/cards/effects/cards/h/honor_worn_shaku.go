package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "babeeaf6-0fe4-491f-bbf5-63a0568b3d6e",
		Name:         "Honor-Worn Shaku",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			UntapSelf: true,
			Label:     "Tap an untapped legendary permanent you control: Untap this artifact.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.TapAnUntapped("an untapped legendary permanent you control", effects.Legendary()),
			Effect:    effects.UntapThisPermanentEffect,
		}},
	})
}
