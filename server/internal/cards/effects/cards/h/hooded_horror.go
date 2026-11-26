package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hooded Horror — Creature — Horror {4}{B}, 4/4 (Commander 2013, S82 older
// sets):
//
//	"This creature can't be blocked as long as defending player controls the
//	 most creatures or is tied for the most."
//
// Scrapdiver Serpent's shape (a CantBeBlockedWhile rule on itself): the
// condition is read of the attacker as blockers are declared, with the
// defending player the one it attacks, or the controller of the planeswalker
// or battle it attacks. Tied for the most counts every other player still in
// the game.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8267561e-bc25-4aaa-8242-f6d7ec88143e",
		Name:         "Hooded Horror",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedWhile(effects.OnSelf(), effects.OldB384DefendingPlayerControlsTheMostCreatures),
		},
	})
}
