package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "6eb5193b-5ef0-4261-acdc-958e161c8118",
		Name:         "Wizened Mentor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggersOnlyOnceEachTurn(WheneverAnOpponentActivates(
				"Wizened Mentor — create a 1/1 white Zombie creature token",
				func(*game.Game, uuid.UUID, game.Card) bool { return true },
				false,
				func(uuid.UUID) Effect { return Do(CreateToken{Template: TokenCard("1/1 white Zombie"), N: 1}) })),
		},
	})
}
