package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zenith Flare — Instant {2}{R}{W} (Ikoria: Lair of Behemoths, #5512):
//
//	"Zenith Flare deals X damage to any target and you gain X life, where X is
//	 the number of cards with a cycling ability in your graveyard."
//
// X is counted once as the spell resolves, from your graveyard as it is then
// (Zenith Flare is still on the stack, so it is not counted, and has no
// cycling anyway). Typecycling counts (CR 702.29e). The life is gained even if
// the damage is prevented; an illegal target means neither happens (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae005f00-3817-40ca-b8be-069a4085cfee",
		Name:         "Zenith Flare",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			x := len(effects.GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
				return effects.GapB158HasCyclingAbility(ctx.Game, ctx.Controller(), c)
			}))
			if x <= 0 {
				return nil
			}
			if err := (effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: x}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Amount: x}.Apply(ctx)
		},
	})
}
