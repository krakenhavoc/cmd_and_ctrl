package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gnaw to the Bone — Instant {2}{G} (The List):
//
//	"You gain 2 life for each creature card in your graveyard.
//	 Flashback {2}{G} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// The count is taken as the spell resolves, from the caster's graveyard
// only. A flashed-back Gnaw is on the stack, not in the graveyard, so it
// is never counted (and is not a creature card anyway). With no creature
// cards no life is gained. Abandon the Post's flashback seam.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "7b92ac54-3883-48cf-b01e-be9797dc1649",
		Name:             "Gnaw to the Bone",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{2}{G}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := 0
			if p := ctx.Game.PlayerByIDForEffect(ctx.Controller()); p != nil && p.Graveyard != nil {
				for _, c := range p.Graveyard.Cards {
					if c.IsCreature() {
						n++
					}
				}
			}
			if n == 0 {
				return nil
			}
			return effects.GainLife{Player: ctx.Controller(), Amount: 2 * n}.Apply(ctx)
		},
	})
}
