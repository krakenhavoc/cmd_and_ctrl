package effects

import (
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
	Register(Spec{
		OracleID:     "57e4a1c0-4218-4a07-88e7-d6083884c2b6",
		Name:         "Champions from Beyond",
		Completeness: CompletenessFull,
		XMatters:     true,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Key:       "Champions from Beyond — create X 1/1 Hero tokens",
				XFromCast: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return CreateToken{Template: TokenCard(jobSelectHeroToken), N: ctx.X()}.Apply(ctx)
				},
			},
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && b16PlayerAttackedWithAtLeast(ev, source, g, 4, lightParty)
			}, lightParty,
				func(g *game.Game, item *game.StackItem) error {
					// The draw waits for the scry's answer, so the card
					// left on top is the card drawn.
					return Scry{N: 2, Then: func(g *game.Game) error {
						return g.DrawNForEffect(item.Controller, 1)
					}}.Apply(NewContext(g, item))
				})),
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && b16PlayerAttackedWithAtLeast(ev, source, g, 8, fullParty)
			}, fullParty, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item).asGroupMember()
				for _, id := range b13AttackingCreaturesYouControl(g, item.Controller) {
					if err := (BoostUntilEOT{Target: id, Power: 4, Toughness: 4, Label: "Full Party — +4/+4"}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})),
		},
	})
}
