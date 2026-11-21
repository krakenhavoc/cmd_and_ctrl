package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// End the Festivities — Sorcery {R} (EDHREC rank 2249):
//
//	"End the Festivities deals 1 damage to each opponent and each
//	 creature and planeswalker they control."
//
// The one-mana one-sided sweeper: every opponent takes one, and so
// does every creature and planeswalker they control — a 1/1 token
// army dies at the state-based sweep, a planeswalker loses a
// loyalty counter (CR 120.3c), and the caster's own board is
// untouched. Damage from the spell as its source, so it is red
// noncombat damage and a prevention shield sees it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09618da4-6e2d-4047-96d7-0576c0754e7d",
		Name:         "End the Festivities",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B21DamageEachOpponentAndTheirCreaturesAndWalkers(ctx, 1)
		},
	})
}
