package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wizened Mentor — Creature — Zombie Cleric {1}{W}, 2/2:
//
//	"Whenever an opponent activates an ability of a permanent that isn't a
//	 mana ability, you create a 1/1 white Zombie creature token. This
//	 ability triggers only once each turn."
//
// Harsh Mentor's watcher over every permanent on the battlefield (the
// activation event, mana abilities excluded by not watching their kind),
// wrapped in TriggersOnlyOnceEachTurn, which tallies per object when the
// trigger goes on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6eb5193b-5ef0-4261-acdc-958e161c8118",
		Name:         "Wizened Mentor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggersOnlyOnceEachTurn(effects.WheneverAnOpponentActivates(
				"Wizened Mentor — create a 1/1 white Zombie creature token",
				func(*game.Game, uuid.UUID, game.Card) bool { return true },
				false,
				func(uuid.UUID) effects.Effect {
					return effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Zombie"), N: 1})
				})),
		},
	})
}
