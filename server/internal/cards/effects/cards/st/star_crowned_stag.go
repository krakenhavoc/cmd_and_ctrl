package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Star-Crowned Stag — Creature — Elk {3}{W}, 3/3 (Core Set 2019, S82 OLD batch 199, #5527):
//
//	"Whenever this creature attacks, tap target creature defending player
//	 controls."
//
// Sidar Jabari's attack trigger (DistractingGeistTap): the defending player is
// read off the Stag's own attack, so in a multiplayer game only a creature of
// the player it attacks can be chosen (CR 506.2a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a375b739-6a6c-4e07-9b1c-a304f84ea2cc",
		Name:         "Star-Crowned Stag",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.DistractingGeistTap("Star-Crowned Stag — tap target creature defending player controls"),
		},
	})
}
