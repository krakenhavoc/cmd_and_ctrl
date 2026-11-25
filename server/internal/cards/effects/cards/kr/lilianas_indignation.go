package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Liliana's Indignation — Sorcery {X}{B} (Shadows over Innistrad, #5543):
//
//	"Mill X cards. Target player loses 2 life for each creature card put
//	 into your graveyard this way."
//
// The caster mills X; only the cards that actually landed in the graveyard
// count (a replacement that sent one elsewhere is not counted, CR 400.7),
// and a creature card counts by its printed types there. The loss is a life
// loss, not damage. If the target is gone by resolution the spell is
// countered on resolution (CR 608.2b) and nothing is milled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c724dd2e-c62d-4a20-a1b9-54264d0b811a",
		Name:         "Liliana's Indignation",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.MillToZone{N: ctx.X(), ThenTail: ctx.ItemTail(lilianasIndignationLoss)}.Apply(ctx)
		},
	})
}

// lilianasIndignationLoss is "target player loses 2 life for each creature
// card put into your graveyard this way", run once the mill has landed and
// told the cards that really reached the graveyard. An on-disk key.
var lilianasIndignationLoss = effects.CardTail("lilianas-indignation/loss", func(ctx *effects.Context, o game.TailOutcome) error {
	creatures := 0
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsCreature() {
			creatures++
		}
	}
	if creatures == 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -2*creatures)
	}
	return nil
})
