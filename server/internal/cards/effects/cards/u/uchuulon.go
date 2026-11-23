package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Uchuulon — {3}{B} Creature — Crab Ooze Horror */4
//
//	"Uchuulon's power is equal to the number of Crabs, Oozes, and/or Horrors
//	 you control.
//	 Horrific Symbiosis — At the beginning of your end step, exile up to one
//	 target creature card from an opponent's graveyard. If you do, create a
//	 token that's a copy of this creature."
//
// The power is a characteristic-defining ability (CR 604.3, layer 7a) counting
// every permanent you control with one of the three creature types, Uchuulon
// itself and tokens included. The exile is "up to one" target, and the token
// is made only if a card was really exiled (CR 608.2c); if Uchuulon has left
// by then the token is a copy of it as it last existed (CR 608.2h).
func init() {
	power := game.StaticAbility{
		Label:     "Uchuulon's power is equal to the number of Crabs, Oozes, and/or Horrors you control.",
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7A_CDA,
		AppliesTo: effects.SelfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := 0
			for i := range g.Battlefield.Cards {
				p := &g.Battlefield.Cards[i]
				if p.Controller == source.Controller && (p.HasSubtypeWord("Crab") || p.HasSubtypeWord("Ooze") || p.HasSubtypeWord("Horror")) {
					n++
				}
			}
			c.Power = n
		},
	}
	effects.Register(effects.Spec{
		OracleID:     "1f266fa4-6667-49db-9d24-dfbabf1449ca",
		Name:         "Uchuulon",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{power},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtYourEndStep("Horrific Symbiosis — exile up to one target creature card from an opponent's graveyard; if you do, create a token that's a copy of this creature",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.ExileTarget{Target: id, ThenTail: ctx.ItemTail(uchuulonCopy)}.Apply(ctx)
					}),
				effects.TargetCardInGraveyard("up to one target creature card from an opponent's graveyard", effects.Creature(), effects.Not(effects.YouOwn())).WithCount(0, 1)),
		},
	})
}

// uchuulonCopy is "if you do, create a token that's a copy of this creature".
// An on-disk key: never renamed, never reused.
var uchuulonCopy = effects.CardTail("tail/uchuulon-copy", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) {
		return nil
	}
	return effects.CreateTokenCopy{Controller: ctx.Controller(), Copy: ctx.Source(), N: 1}.Apply(ctx)
})
