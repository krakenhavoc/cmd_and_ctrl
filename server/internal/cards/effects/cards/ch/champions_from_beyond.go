package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Champions from Beyond — Enchantment {X}{W}{W}:
//
//	"When this enchantment enters, create X 1/1 colorless Hero creature
//	 tokens.
//	 Light Party — Whenever you attack with four or more creatures, scry
//	 2, then draw a card.
//	 Full Party — Whenever you attack with eight or more creatures,
//	 those creatures get +4/+4 until end of turn."
//
// X is the cast's X, stamped on the entry trigger by the engine
// (XFromCast, as Arboreal Alliance's is). The party triggers are
// Aurelia's "attacks with N or more" condition, once per declaration
// (OncePerBatch), for the Champions' controller only. Full Party pumps
// the creatures you attacked with, read as the ability resolves
// (CR 611.2c).
//
// No simplification.
func init() {
	const (
		lightParty = "Champions from Beyond — Light Party: scry 2, then draw a card"
		fullParty  = "Champions from Beyond — Full Party: attackers get +4/+4 until end of turn"
	)
	effects.Register(effects.Spec{
		OracleID:     "57e4a1c0-4218-4a07-88e7-d6083884c2b6",
		Name:         "Champions from Beyond",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: effects.Self,
				Key:       "Champions from Beyond — create X 1/1 Hero tokens",
				XFromCast: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.CreateToken{Template: effects.TokenCard(effects.JobSelectHeroToken), N: ctx.X()}.Apply(ctx)
				},
			},
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && effects.B16PlayerAttackedWithAtLeast(ev, source, g, 4, lightParty)
			}, lightParty,
				func(g *game.Game, item *game.StackItem) error {
					// The draw waits for the scry's answer, so the card
					// left on top is the card drawn.
					ctx := effects.NewContext(g, item)
					return effects.Scry{N: 2, ThenTail: ctx.TailN(effects.ThenDrawCards, 1)}.Apply(ctx)
				})),
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && effects.B16PlayerAttackedWithAtLeast(ev, source, g, 8, fullParty)
			}, fullParty, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item).AsGroupMember()
				for _, id := range effects.B13AttackingCreaturesYouControl(g, item.Controller) {
					if err := (effects.BoostUntilEOT{Target: id, Power: 4, Toughness: 4, Label: "Full Party — +4/+4"}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})),
		},
	})
}
