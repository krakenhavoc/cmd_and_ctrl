package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ajani Unrelenting — Legendary Planeswalker — Ajani {4}{R}{R}, starting
// loyalty 5:
//
//	"Whenever you activate a loyalty ability, create a 2/2 colorless
//	 Wizard Soldier creature token named Cadet.
//	 +1: Creatures you control get +1/+0 and gain haste until end of turn.
//	 −2: Discard your hand, then draw a card for each creature you control.
//	 −3: Ajani deals 4 damage to each creature except for tokens you
//	 control."
//
// The trigger watches the activation event, so it goes on the stack above
// the loyalty ability and the Cadet is already there when the ability
// resolves: the −2 counts it, and the −3 spares it. "Whenever you
// activate a loyalty ability" is any loyalty ability of any planeswalker
// you control, Ajani's own included. The +1 is ONE effect (one timestamp)
// on the creatures you control as it resolves (CR 611.2c), so a creature
// that arrives later gets neither the bonus nor haste. The −2 discards
// the whole hand first, then counts the creatures. The −3 is simultaneous
// damage to every creature except the tokens its controller controls: an
// opponent's tokens are hit, your nontoken creatures are hit.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "71620e92-45a4-49cb-94d5-348013ecf886",
		Name:            "Ajani Unrelenting",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventActivateAbility, youActivatedALoyaltyAbility,
				"Ajani Unrelenting — create a 2/2 Wizard Soldier token named Cadet",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{
						Controller: item.Controller,
						Template:   effects.TokenCard("2/2 colorless Wizard Soldier named Cadet"),
						N:          1,
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Creatures you control get +1/+0 and gain haste until end of turn.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.UntilEndOfTurn(effects.NewContext(g, item), uuid.Nil, effects.And(effects.Creature(), effects.YouControl()),
						"Ajani Unrelenting — +1/+0 and haste", game.ModifyPTMod(1, 0), game.AddKeywordsMod("haste"))
				},
			},
			{
				Label: "−2: Discard your hand, then draw a card for each creature you control.",
				Cost:  effects.LoyaltyCost(-2),
				// The draw waits for a prompt a discard asks (a replacement offer such as Library of Leng's; CR 608.2c, #4155)
				// and counts the creatures when it runs.
				Effect: effects.Do(
					effects.Instruction(func(ctx *effects.Context) error {
						_, err := effects.DiscardWholeHand(ctx.Game, ctx.Controller())
						return err
					}),
					effects.Instruction(func(ctx *effects.Context) error {
						n := 0
						for _, c := range ctx.Game.BattlefieldCardsForEffect() {
							if c.Controller == ctx.Controller() && c.IsCreature() {
								n++
							}
						}
						return effects.DrawCards{N: n}.Apply(ctx)
					}),
				),
			},
			{
				Label:   "−3: Ajani deals 4 damage to each creature except for tokens you control.",
				Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 4, Partial: true}},
				Cost:    effects.LoyaltyCost(-3),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DamageEachMatching(effects.NewContext(g, item),
						effects.And(effects.Creature(), effects.Not(effects.And(effects.IsTokenPredicate(), effects.YouControl()))), 4)
				},
			},
		},
	})
}

// youActivatedALoyaltyAbility is "whenever you activate a loyalty
// ability": the activation event's Loyalty bit (CR 606.2), by the
// source's controller. A copy of an ability is not activated (CR 707.10),
// so it never matches.
func youActivatedALoyaltyAbility(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventActivateAbility && ev.Loyalty && ev.Actor == source.Controller
}
