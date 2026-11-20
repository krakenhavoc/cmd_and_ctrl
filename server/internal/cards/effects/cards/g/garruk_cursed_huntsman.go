package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Garruk, Cursed Huntsman — Legendary Planeswalker — Garruk {4}{B}{G},
// starting loyalty 5:
//
//	"0: Create two 2/2 black and green Wolf creature tokens with "When this
//	 token dies, put a loyalty counter on each Garruk you control."
//	 −3: Destroy target creature. Draw a card.
//	 −6: You get an emblem with "Creatures you control get +3/+3 and have
//	 trample.""
//
// The Wolf is a catalog token (cmdr2S51WolfToken) whose dies trigger puts
// the counter on each Garruk its controller controls as the trigger
// resolves. The emblem is Elspeth, Sun's Champion's shape.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e0cef79c-ad47-4cbc-9d73-a913e487ccb7",
		Name:            "Garruk, Cursed Huntsman",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Emblem: &effects.EmblemSpec{
			Label: "Garruk, Cursed Huntsman emblem",
			Text:  "Creatures you control get +3/+3 and have trample.",
			Static: []game.StaticAbility{
				{
					Layer:     game.Layer7PT,
					SubLayer:  game.SubLayer7C_Modify,
					AppliesTo: effects.CreaturesTheEmblemsOwnerControls,
					Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
						c.Power += 3
						c.Toughness += 3
					},
				},
				effects.KeywordGrant(effects.CreaturesTheEmblemsOwnerControls, "trample"),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "0: Create two 2/2 black and green Wolf creature tokens with \"When this token dies, put a loyalty counter on each Garruk you control.\"",
				Cost:    effects.LoyaltyCost(0),
				Purpose: game.Purpose{Tokens: 2},
				Effect:  effects.Do(effects.CreateToken{Template: effects.Cmdr2S51WolfToken(), N: 2}),
			},
			{
				Label:   "−3: Destroy target creature. Draw a card.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCreature("target creature"),
				Purpose: game.Purpose{Draws: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					if err := (effects.DestroyTarget{Target: ts[0].ID}).Apply(ctx); err != nil {
						return err
					}
					return effects.DrawCards{N: 1}.Apply(ctx)
				},
			},
			{
				Label:  "−6: You get an emblem with \"Creatures you control get +3/+3 and have trample.\"",
				Cost:   effects.LoyaltyCost(-6),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}
