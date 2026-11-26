package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sarkhan the Mad — Legendary Planeswalker — Sarkhan {3}{B}{R}, loyalty 7
// (Rise of the Eldrazi, S82 OLD b486):
//
//	"0: Reveal the top card of your library and put it into your hand.
//	   Sarkhan deals damage to himself equal to that card's mana value.
//	 −2: Target creature's controller sacrifices it, then that player creates
//	   a 5/5 red Dragon creature token with flying.
//	 −4: Each Dragon creature you control deals damage equal to its power to
//	   target player or planeswalker."
//
// The 0 is Dark Confidant's flip: the card is revealed (not drawn), its mana
// value read before it moves, and Sarkhan is dealt that much damage by himself,
// which removes that much loyalty (CR 120.3c); a land or an empty library costs
// him nothing. The −2 reads the creature's controller before the sacrifice; a
// sacrifice that did not happen (a replacement, CR 701.21a) still makes the
// token, which is a separate instruction, and it is that player's, not the
// activator's. The −4 deals each Dragon's damage from that Dragon (so deathtouch
// and lifelink see each), together as one damage instance, with its power read as
// the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e70a93e4-2fe7-46d3-8958-153ab2fcfd99",
		Name:            "Sarkhan the Mad",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 7,
		Activated: []effects.ActivatedAbility{
			{
				Label: "0: Reveal the top card of your library and put it into your hand. Sarkhan deals damage to himself equal to that card's mana value.",
				Cost:  effects.LoyaltyCost(0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					var revealed []uuid.UUID
					if err := (effects.RevealTopOfLibrary{
						Player:   item.Controller,
						N:        1,
						Reason:   "Sarkhan the Mad — reveal the top card of your library",
						Revealed: &revealed,
					}).Apply(ctx); err != nil {
						return err
					}
					if len(revealed) == 0 {
						return nil
					}
					mv := 0
					if c, ok := g.LookupCardForEffect(revealed[0]); ok {
						mv = c.ManaValue()
					}
					return effects.BounceToHand{Target: revealed[0], ThenTail: ctx.TailN(sarkhanTheMadSelfDamage, mv)}.Apply(ctx)
				},
			},
			{
				Label:   "−2: Target creature's controller sacrifices it, then that player creates a 5/5 red Dragon creature token with flying.",
				Cost:    effects.LoyaltyCost(-2),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						victim, ok := g.LookupCardForEffect(t.ID)
						if !ok {
							return nil
						}
						then := game.KeyedTail{Ref: sarkhanTheMadDragon, Player: victim.Controller, Source: ctx.Source()}
						return effects.SacrificePermanent{Target: t.ID, ThenTail: then}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label:   "−4: Each Dragon creature you control deals damage equal to its power to target player or planeswalker.",
				Cost:    effects.LoyaltyCost(-4),
				Targets: effects.TargetPlayerOrPlaneswalker(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					targets := ctx.LegalTargets()
					if len(targets) == 0 {
						return nil
					}
					dragon := ctx.TextSubtype("Dragon")
					dragons := effects.MatchingBattlefield(ctx, effects.And(effects.YouControl(), effects.Creature(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.HasSubtype(dragon)
					}))
					if len(dragons) == 0 {
						return nil
					}
					return g.DamageInstanceForEffect(func() error {
						for _, d := range dragons {
							if err := (effects.DealDamage{Source: d.InstanceID, Target: targets[0].ID, Amount: d.CurrentPower()}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					})
				},
			},
		},
	})
}

// sarkhanTheMadDragon is "then that player creates a 5/5 red Dragon creature
// token with flying", run once the sacrifice has been settled: the tail's
// player is the creature's controller. An on-disk key: never renamed, never
// reused.
var sarkhanTheMadDragon = effects.CardTail("then/sarkhan-the-mad-dragon", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.CreateToken{Controller: o.Tail.Player, Template: effects.TokenCard("5/5 red Dragon with flying"), N: 1}.Apply(ctx)
})

// sarkhanTheMadSelfDamage is "Sarkhan deals damage to himself equal to that
// card's mana value", run once the card has been put into the hand (or sent
// elsewhere: the damage is its own instruction, not an "if you do"). The mana
// value was read as the card was revealed and rides the tail as N. An on-disk
// key: never renamed, never reused.
var sarkhanTheMadSelfDamage = effects.CardTail("then/sarkhan-the-mad-self-damage", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.DealDamage{Source: o.Tail.Source, Target: o.Tail.Source, Amount: o.Tail.N}.Apply(ctx)
})
