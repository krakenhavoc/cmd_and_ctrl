package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grave Strength — Sorcery {1}{B} (Foundations Jumpstart):
//
//	"Choose target creature. Mill three cards, then put a +1/+1 counter
//	 on that creature for each creature card in your graveyard."
//
// The mill goes first and the count is taken after it (Deathbonnet
// Sprout's creatureCardsInGraveyard), so the milled creatures count. A
// target that is illegal as the spell resolves makes the whole spell do
// nothing, the mill included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "feb2f08e-79c7-4d51-bf55-e649f5ff5e07",
		Name:         "Grave Strength",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.MillCards{N: 3}).Apply(ctx); err != nil {
				return err
			}
			n := effects.CreatureCardsInGraveyard(ctx.Game, ctx.Controller())
			if n == 0 {
				return nil
			}
			return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
		},
	})
}
