package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rith's Charm — Instant {R}{G}{W}:
//
//	"Choose one —
//	 • Destroy target nonbasic land.
//	 • Create three 1/1 green Saproling creature tokens.
//	 • Prevent all damage a source of your choice would deal this turn."
//
// ADR 0108 §7 (#1904): the third mode is the shield against a source
// chosen as the charm resolves (CR 609.7a), preventing all of its damage,
// to anything, for the rest of the turn. The first two modes are the
// catalog's destroy and token primitives.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9dbccdc5-61f7-4a7d-bbbc-5cab393e24f7",
		Name:         "Rith's Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Destroy target nonbasic land.", effects.TargetPermanent("target nonbasic land", effects.NonbasicLand())),
			effects.ModeWithPurpose(effects.Mode("Create three 1/1 green Saproling creature tokens."), game.Purpose{Tokens: 3}),
			effects.Mode("Prevent all damage a source of your choice would deal this turn."),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			switch {
			case ctx.HasMode(0):
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
					}
				}
				return nil
			case ctx.HasMode(1):
				return effects.CreateToken{Controller: item.Controller, Template: effects.B11GreenSaprolingToken(), N: 3}.Apply(ctx)
			case ctx.HasMode(2):
				return effects.PreventDamageFromChosenSource(effects.ShieldAnything).Apply(ctx)
			}
			return nil
		},
	})
}
