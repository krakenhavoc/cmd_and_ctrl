package effects

import (
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
	Register(Spec{
		OracleID:        "e0cef79c-ad47-4cbc-9d73-a913e487ccb7",
		Name:            "Garruk, Cursed Huntsman",
		Completeness:    CompletenessFull,
		StartingLoyalty: 5,
		Emblem: &EmblemSpec{
			Label: "Garruk, Cursed Huntsman emblem",
			Text:  "Creatures you control get +3/+3 and have trample.",
			Static: []game.StaticAbility{
				{
					Layer:     game.Layer7PT,
					SubLayer:  game.SubLayer7C_Modify,
					AppliesTo: creaturesTheEmblemsOwnerControls,
					Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
						c.Power += 3
						c.Toughness += 3
					},
				},
				KeywordGrant(creaturesTheEmblemsOwnerControls, "trample"),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label:   "0: Create two 2/2 black and green Wolf creature tokens with \"When this token dies, put a loyalty counter on each Garruk you control.\"",
				Cost:    LoyaltyCost(0),
				Purpose: game.Purpose{Tokens: 2},
				Effect:  Do(CreateToken{Template: cmdr2S51WolfToken(), N: 2}),
			},
			{
				Label:   "−3: Destroy target creature. Draw a card.",
				Cost:    LoyaltyCost(-3),
				Targets: TargetCreature("target creature"),
				Purpose: game.Purpose{Draws: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					if err := (DestroyTarget{Target: ts[0].ID}).Apply(ctx); err != nil {
						return err
					}
					return DrawCards{N: 1}.Apply(ctx)
				},
			},
			{
				Label:  "−6: You get an emblem with \"Creatures you control get +3/+3 and have trample.\"",
				Cost:   LoyaltyCost(-6),
				Effect: Do(CreateEmblem{}),
			},
		},
	})
}
