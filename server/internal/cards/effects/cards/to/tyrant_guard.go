package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tyrant Guard — Creature — Tyranid {X}{2}{G}, 3/3:
//
//	"Ravenous (This creature enters with X +1/+1 counters on it. If X is 5 or
//	 more, draw a card when it enters.)
//	 Shieldwall — Sacrifice this creature: Creatures you control with
//	 counters on them gain hexproof and indestructible until end of turn."
//
// The creatures affected are those with a counter as the ability resolves
// (CR 611.2c); the Guard itself has been sacrificed as the cost.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "560e6fd6-c061-44ba-b5c5-814d6a14e469",
		Name:                       "Tyrant Guard",
		Completeness:               effects.CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: effects.RavenousEntry(),
		Triggered:                  []game.TriggeredAbility{effects.RavenousDraw("Tyrant Guard")},
		Activated: []effects.ActivatedAbility{{
			Label:   "Shieldwall — Sacrifice this creature: Creatures you control with counters on them gain hexproof and indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect | game.AnswerSurvive},
			Cost:    effects.SacrificeThis(),
			Effect:  effects.TyranidWithCountersHaveHexproofAndIndestructible("Tyrant Guard — Shieldwall"),
		}},
	})
}
