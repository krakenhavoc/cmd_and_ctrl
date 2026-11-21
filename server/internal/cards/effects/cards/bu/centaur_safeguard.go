package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Centaur Safeguard — Creature — Centaur Warrior {2}{G/W}, 3/1 (Ravnica:
// City of Guilds, #5600):
//
//	"({G/W} can be paid with either {G} or {W}.)
//	 When this creature dies, you may gain 3 life."
//
// The dies trigger goes on the stack and asks its "you may" as it resolves
// (CR 603.5); the hybrid pip is printed cost data (CR 107.4e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "468604f2-52e6-4bb1-8a4f-ad0851288cba",
		Name:         "Centaur Safeguard",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WhenThisDies("Centaur Safeguard — you may gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
				"Centaur Safeguard — gain 3 life?"),
		},
	})
}
