package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wondrous Crucible — Artifact {7} (The Brothers' War Commander):
//
//	"Permanents you control have ward {2}.
//	 At the beginning of your end step, mill two cards, then exile a
//	 nonland card at random from your graveyard. Copy it. You may cast
//	 the copy without paying its mana cost. (A copy of a permanent spell
//	 becomes a token.)"
//
// The ward grant is WardGranted over every permanent the controller has,
// the Crucible included. The end-step trigger mills first (the milled
// cards are in the graveyard the random pick reads), then picks through
// the keyed random stream and exiles; CopyAndCast makes the copy of the
// card that REALLY reached exile and offers the free cast while the
// trigger is still resolving (CR 707.12, 608.2g). A graveyard with no
// nonland card does nothing after the mill.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d0e80414-abde-4f5e-9539-3efc83aa507f",
		Name:         "Wondrous Crucible",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WardGranted(effects.WardMana("{2}"), "Wondrous Crucible — permanents you control have ward {2}",
				func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Controller == source.Controller
				}),
			effects.AtYourEndStep("Wondrous Crucible — mill two cards, then exile a nonland card at random from your graveyard, copy it, and you may cast the copy without paying its mana cost",
				wondrousCrucibleResolve),
		},
	})
}

// wondrousCrucibleResolve is the end-step trigger's body.
func wondrousCrucibleResolve(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.MillCards{N: 2}).Apply(ctx); err != nil {
		return err
	}
	pick := effects.RandomPick(ctx, effects.GraveyardIDs(g, ctx.Controller(), func(c game.Card) bool {
		return !c.IsLand()
	}), 1)
	if len(pick) == 0 {
		return nil
	}
	return g.ExileCardKeyedForEffect(pick[0], ctx.ItemTail(wondrousCrucibleExiled))
}

// wondrousCrucibleExiled is "copy it, and you may cast the copy" for the
// card that reached exile. An on-disk key.
var wondrousCrucibleExiled = effects.CardTail("wondrous-crucible/copy-and-cast", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.CopyAndCast{Cards: o.IDs, Free: true}.Apply(ctx)
})
