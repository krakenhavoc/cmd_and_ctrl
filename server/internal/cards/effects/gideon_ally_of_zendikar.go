package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gideon, Ally of Zendikar — Legendary Planeswalker — Gideon
// {2}{W}{W}, starting loyalty 4:
//
//	"+1: Until end of turn, Gideon, Ally of Zendikar becomes a 5/5
//	 Human Soldier Ally creature with indestructible that's still a
//	 planeswalker. Prevent all damage that would be dealt to him this
//	 turn.
//	 0: Create a 2/2 white Knight Ally creature token.
//	 −4: You get an emblem with "Creatures you control get +1/+1.""
//
// THE +1 is animateGideon (gideon_animate.go): one data record that makes
// him a creature that is still a planeswalker until end of turn, and the
// source shield that stops damage to him this turn. Damage that can't be
// prevented (CR 615.12) removes loyalty and is marked on him, and both
// state-based actions apply (#2046, ADR 0032 amendment of 2026-10-07).
// THE 0 is a plain token. THE EMBLEM is an ordinary layer-7c anthem
// owned by the player who got it (CR 114.2), the Elspeth, Sun's
// Champion shape.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "51eae7ff-fed9-4afe-8053-7690379449dd",
		Name:         "Gideon, Ally of Zendikar",
		Completeness: CompletenessFull,
		// The fallback for tokens, fixtures and the dev spawner; an
		// imported deck reads printed loyalty (ADR 0032 §1).
		StartingLoyalty: 4,
		Emblem: &EmblemSpec{
			Label: "Gideon, Ally of Zendikar emblem",
			Text:  "Creatures you control get +1/+1.",
			Static: []game.StaticAbility{{
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: creaturesTheEmblemsOwnerControls,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
					c.Toughness++
				},
			}},
		},
		Activated: []ActivatedAbility{
			{
				Label: "+1: Until end of turn, Gideon, Ally of Zendikar becomes a 5/5 Human Soldier Ally creature with indestructible that's still a planeswalker. Prevent all damage that would be dealt to him this turn.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return animateGideon(g, item, gideonAnimation{
						Label:          "Gideon, Ally of Zendikar — a 5/5 Human Soldier Ally creature with indestructible until end of turn",
						Subtypes:       []string{"Human", "Soldier", "Ally"},
						Power:          5,
						Toughness:      5,
						Indestructible: true,
					})
				},
			},
			{
				Label: "0: Create a 2/2 white Knight Ally creature token.",
				Cost:  LoyaltyCost(0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{
						Controller: item.Controller,
						Template:   TokenCard("2/2 white Knight Ally"),
						N:          1,
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "−4: You get an emblem with \"Creatures you control get +1/+1.\"",
				Cost:  LoyaltyCost(-4),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
