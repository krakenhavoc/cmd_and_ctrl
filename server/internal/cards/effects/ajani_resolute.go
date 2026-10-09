package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ajani Resolute — Legendary Planeswalker — Ajani {1}{W}, starting
// loyalty 2:
//
//	"Whenever you gain life, put a loyalty counter on Ajani.
//	 0: You gain 1 life.
//	 −4: Create a 2/2 white Cat Soldier creature token named Ajani's
//	 Pridemate with "Whenever you gain life, put a +1/+1 counter on this
//	 token."
//	 −10: You get an emblem with "Creatures you control get +2/+2.""
//
// The trigger is one per life-gain event, so the 0 ticks him up by one
// each time and a lifelink hit does too. It puts the counter on Ajani
// only while he is still on the battlefield. The Pridemate is a catalog
// token (slug "ajanis-pridemate") carrying its own gain-life trigger.
// The emblem is a layer 7c anthem owned by the player who made it.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "0d60c095-4c6c-4618-8ef3-508f3d454efa",
		Name:            "Ajani Resolute",
		Completeness:    CompletenessFull,
		StartingLoyalty: 2,
		Emblem: &EmblemSpec{
			Label: "Ajani Resolute emblem",
			Text:  "Creatures you control get +2/+2.",
			Static: []game.StaticAbility{{
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: creaturesTheEmblemsOwnerControls,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power += 2
					c.Toughness += 2
				},
			}},
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Ajani Resolute — put a loyalty counter on him",
				func(g *game.Game, item *game.StackItem) error {
					if !b09SourceStillOnBattlefield(g, item) {
						return nil
					}
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterLoyalty, N: 1}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{
			{
				Label:  "0: You gain 1 life.",
				Cost:   LoyaltyCost(0),
				Effect: Do(GainLife{Amount: 1}),
			},
			{
				Label: "−4: Create a 2/2 white Cat Soldier creature token named Ajani's Pridemate with \"Whenever you gain life, put a +1/+1 counter on this token.\"",
				Cost:  LoyaltyCost(-4),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Controller: item.Controller, Template: AjanisPridemateToken(), N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "−10: You get an emblem with \"Creatures you control get +2/+2.\"",
				Cost:  LoyaltyCost(-10),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
