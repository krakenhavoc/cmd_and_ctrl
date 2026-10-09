package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:        "71620e92-45a4-49cb-94d5-348013ecf886",
		Name:            "Ajani Unrelenting",
		Completeness:    CompletenessFull,
		StartingLoyalty: 5,
		Triggered: []game.TriggeredAbility{
			On(game.EventActivateAbility, youActivatedALoyaltyAbility,
				"Ajani Unrelenting — create a 2/2 Wizard Soldier token named Cadet",
				func(g *game.Game, item *game.StackItem) error {
					return CreateToken{
						Controller: item.Controller,
						Template:   TokenCard("2/2 colorless Wizard Soldier named Cadet"),
						N:          1,
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{
			{
				Label: "+1: Creatures you control get +1/+0 and gain haste until end of turn.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return untilEndOfTurn(NewContext(g, item), uuid.Nil, And(Creature(), YouControl()),
						"Ajani Unrelenting — +1/+0 and haste", game.ModifyPTMod(1, 0), game.AddKeywordsMod("haste"))
				},
			},
			{
				Label: "−2: Discard your hand, then draw a card for each creature you control.",
				Cost:  LoyaltyCost(-2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if _, err := discardWholeHand(g, item.Controller); err != nil {
						return err
					}
					n := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() {
							n++
						}
					}
					return DrawCards{Player: item.Controller, N: n}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "−3: Ajani deals 4 damage to each creature except for tokens you control.",
				Cost:  LoyaltyCost(-3),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return damageEachMatching(NewContext(g, item),
						And(Creature(), Not(And(IsTokenPredicate(), YouControl()))), 4)
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
