package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wandering Goblins — Creature — Goblin Warrior {2}{R}, 0/3 (Conflux,
// #5587):
//
//	"Domain — {3}: This creature gets +1/+0 until end of turn for each
//	 basic land type among lands you control."
//
// Sunbathing Rootwalla's row without the once-a-turn limit: the count is
// the shared DomainCount, fixed as the ability resolves (CR 608.2h), and
// the boost ends in the cleanup step (CR 514.2). No {T} in the cost, so
// it can be activated any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8046981c-d68c-4fbb-a133-8037996c9d10",
		Name:         "Wandering Goblins",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Domain — {3}: This creature gets +1/+0 until end of turn for each basic land type among lands you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{3}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				n := effects.DomainCount(g, item.Controller)
				return effects.BoostUntilEOT{Target: item.SourceCardID, Power: n, Label: "Wandering Goblins — +1/+0 for each basic land type"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
