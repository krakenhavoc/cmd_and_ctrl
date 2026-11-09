package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Edgar, King of Figaro — Legendary Creature — Human Artificer Noble
// {4}{U}{U}, 4/5:
//
//	"When Edgar enters, draw a card for each artifact you control.
//	 Two-Headed Coin — The first time you flip one or more coins each turn,
//	 those coins come up heads and you win those flips."
//
// The enter trigger counts your artifacts as it resolves. Two-Headed Coin
// is the CoinFlipRules seam (game/coin_flip_rules.go): derived off the
// battlefield every time you flip, called or uncalled, for the first flip
// action each turn, with every coin of that flip heads and won.
func init() {
	Register(Spec{
		OracleID:     "264efed9-65ab-40ad-877f-c6c25ca0281e",
		Name:         "Edgar, King of Figaro",
		Completeness: CompletenessFull,
		CoinFlipRules: []game.CoinFlipRule{
			{Label: "Two-Headed Coin", FirstTimeEachTurn: true, HeadsAndWin: true},
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Edgar, King of Figaro — draw a card for each artifact you control",
				func(g *game.Game, item *game.StackItem) error {
					n := len(permanentsControlledByMatching(g, item.Controller, Artifact()))
					if n == 0 {
						return nil
					}
					return DrawCards{N: n}.Apply(NewContext(g, item))
				}),
		},
	})
}
