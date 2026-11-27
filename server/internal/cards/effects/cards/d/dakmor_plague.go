package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dakmor Plague — {3}{B}{B} Sorcery (Starter 1999, S82 OLD batch 702, #5620):
//
//	"Dakmor Plague deals 3 damage to each creature and each player."
//
// Famine's body: B23DamageEachCreatureAndEachPlayer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6cc516d-5a4b-455a-ab16-367d64215527",
		Name:         "Dakmor Plague",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B23DamageEachCreatureAndEachPlayer(ctx, 3)
		},
	})
}
