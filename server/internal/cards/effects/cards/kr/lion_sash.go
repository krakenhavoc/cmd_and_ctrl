package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lion Sash — Artifact Creature — Equipment Cat {1}{W}, 1/1:
//
//	"{W}: Exile target card from a graveyard. If it was a permanent card,
//	 put a +1/+1 counter on this permanent.
//	 Equipped creature gets +1/+1 for each +1/+1 counter on this
//	 Equipment.
//	 Reconfigure {2}"
//
// The exile clause is Cling to Dust's ExileThenIfItWas: the card's type
// is read before it moves, and the counter waits for the exile to
// happen. The counter goes on the Sash only while it is still the same
// object on the battlefield. While the Sash fights on its own its
// counters grow it directly; while attached they pump the host
// (PumpAttachedPer). Reconfigure is reconfigure.go (#2639).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "50330e48-db74-4c5f-a0bc-f9a8607e8f31",
		Name:         "Lion Sash",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, func(_ *game.Game, source *game.Card) int {
				return source.Counters["+1/+1"]
			}),
		},
		Activated: append([]effects.ActivatedAbility{{
			Label:   "{W}: Exile target card from a graveyard. If it was a permanent card, put a +1/+1 counter on this permanent.",
			Cost:    effects.ManaCost("{W}"),
			Targets: effects.TargetCardInGraveyard("target card from a graveyard"),
			// #4559: what the exile takes away and the counter it grows.
			Purpose: effects.ForTargets(effects.ExilesFromGraveyard(0, game.ExileGrowsPermanent)),
			Effect:  lionSashExile,
		}}, effects.Reconfigure("{2}")...),
	})
}

func lionSashExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return effects.ExileThenIfItWas{
			Target: t.ID,
			Was:    func(c game.Card) bool { return c.IsPermanent() },
			Keyed:  lionSashIfPermanent,
		}.Apply(ctx)
	}
	return nil
}

// lionSashIfPermanent is the exile's keyed "if it was a permanent card,
// put a +1/+1 counter on Lion Sash" (#4253 slice 10a).
var lionSashIfPermanent = effects.IfItWasTail("if-it-was/lion-sash-counter", func(ctx *effects.Context) error {
	if !effects.OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	return effects.AddCounter{Target: ctx.Source(), Kind: "+1/+1", N: 1}.Apply(ctx)
}, nil)
