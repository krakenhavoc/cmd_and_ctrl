package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gates Ablaze — Sorcery {2}{R} (Ravnica Allegiance):
//
//	"Gates Ablaze deals X damage to each creature, where X is the number of
//	 Gates you control."
//
// X is counted once as the spell resolves (CR 608.2h) and the damage is dealt to
// every creature at once, Gates Ablaze being the source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b97bf8e-d5c4-4eae-a453-73456e0461a1",
		Name:         "Gates Ablaze",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := effects.CountYouControl(ctx, effects.HasSubtype("Gate"))
			return effects.DamageEachMatching(ctx, effects.Creature(), x)
		},
	})
}
