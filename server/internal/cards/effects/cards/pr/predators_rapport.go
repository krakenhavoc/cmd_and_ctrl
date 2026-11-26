package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Predator's Rapport — Instant {2}{G} (Gatecrash, S82 OLD b408, #5565):
//
//	"Choose target creature you control. You gain life equal to that creature's
//	 power plus its toughness."
//
// One creature target, read as the spell resolves (CR 608.2b: a target that is
// gone makes the spell fizzle, so no life is gained): its power and toughness
// as they are now, layers applied. A negative total gains nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7121b877-d262-4627-acd8-a159f8a779b1",
		Name:         "Predator's Rapport",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, found := ctx.Game.LookupCardForEffect(id)
			if !found || !effects.OnBattlefield(ctx.Game, id) {
				return nil
			}
			n := c.CurrentPower() + c.CurrentToughness()
			if n <= 0 {
				return nil
			}
			return effects.GainLife{Player: ctx.Controller(), Amount: n}.Apply(ctx)
		},
	})
}
