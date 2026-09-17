package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Gold Saucer — "{T}: Add {C}. {2}, {T}: Flip a coin. If you win the
// flip, create a Treasure token. {3}, {T}, Sacrifice two artifacts: Draw a
// card."
func init() {
	Register(Spec{
		OracleID:     "93e38650-ce22-4ab9-b79d-cc7b6477c075",
		Name:         "The Gold Saucer",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true},
			Produced:                "{C}",
			Label:                   "{T}: Add {C}",
			IgnoreCommanderIdentity: true,
		}},
		Activated: []ActivatedAbility{
			{
				Label: "{2}, {T}: Flip a coin. If you win the flip, create a Treasure token.",
				Cost:  Plus(ManaCost("{2}"), TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					controller := item.Controller
					g.FlipCoinForEffect(game.CoinFlipSpec{
						Flipper:  item.Controller,
						Source:   item.SourceCardID,
						Question: "The Gold Saucer — call the coin flip",
						Then: func(g *game.Game, result game.CoinFlipResult) error {
							if len(result.Won) > 0 && result.Won[0] {
								return treasureTokens(g, controller, 1)
							}
							return nil
						},
					})
					return nil
				},
			},
			{
				Label: "{3}, {T}, Sacrifice two artifacts: Draw a card.",
				Cost:  Plus(ManaCost("{3}"), TapCost(), SacrificeN(2, "two artifacts", Artifact())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
