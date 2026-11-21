package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brindle Shoat — Creature — Boar {1}{G}, 1/1 (Magic 2013):
//
//	"When this creature dies, create a 3/3 green Boar creature token."
//
// The token is created by the death trigger, under the controller of the
// trigger (the Shoat's controller as it died).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d385c2e1-a62e-43ad-80ef-a6178deac82c",
		Name:         "Brindle Shoat",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Brindle Shoat — create a 3/3 green Boar creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 green Boar"), N: 1})),
		},
	})
}
