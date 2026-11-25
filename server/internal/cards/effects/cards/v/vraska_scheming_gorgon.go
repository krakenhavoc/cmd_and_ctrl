package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vraska, Scheming Gorgon — Legendary Planeswalker — Vraska {4}{B}{B}, loyalty 5:
//
//	"+2: Creatures you control get +1/+0 until end of turn.
//	 −3: Destroy target creature.
//	 −10: Until end of turn, creatures you control gain deathtouch and
//	 'Whenever this creature deals damage to an opponent, that player loses
//	 the game.'"
//
// +2 is one until-end-of-turn pump over the creatures controlled as it resolves
// (CR 611.2c). −10 is Predators' Hour's shape: a granted keyword and a granted
// triggered ability for the creatures controlled as it resolves, lasting to the
// cleanup step; the trigger fires for any damage the creature deals to an
// opponent, combat or not, and that player loses the game (CR 104.3).
//
// No simplification.
func init() {
	const grant = "vraska-scheming-gorgon/damage-loses-the-game"
	effects.Register(effects.Spec{
		OracleID:        "83045a9c-57da-411f-80aa-56178d14856b",
		Name:            "Vraska, Scheming Gorgon",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsDamageToAnOpponent("Vraska, Scheming Gorgon — that player loses the game",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						return effects.LoseTheGame{Player: ctx.Trigger().Event.Target}.Apply(ctx)
					}),
			},
			Text: "Whenever this creature deals damage to an opponent, that player loses the game.",
		}},
		Activated: []effects.ActivatedAbility{
			{
				Label: "+2: Creatures you control get +1/+0 until end of turn.",
				Cost:  effects.LoyaltyCost(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Label: "Vraska, Scheming Gorgon — +1/+0"}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "−3: Destroy target creature.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DestroyFirstLegalCardTarget(g, item)
				},
			},
			{
				Label: "−10: Until end of turn, creatures you control gain deathtouch and \"Whenever this creature deals damage to an opponent, that player loses the game.\"",
				Cost:  effects.LoyaltyCost(-10),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					yours := effects.And(effects.Creature(), effects.YouControl())
					if err := (effects.GrantKeywordUntilEOT{Match: yours, Keywords: []string{"deathtouch"}, Label: "Vraska, Scheming Gorgon — deathtouch"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantAbilitiesFor{Match: yours, Keys: []string{grant}, Label: "Vraska, Scheming Gorgon — damage to an opponent loses them the game"}.Apply(ctx)
				},
			},
		},
	})
}
