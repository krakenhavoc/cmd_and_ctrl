package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloudblazer — Creature — Human Scout {3}{W}{U}, 2/2 (EDHREC rank
// 3505):
//
//	"Flying
//	 When this creature enters, you gain 2 life and draw two cards."
//
// Mulldrifter with a life gain. The entry trigger gains the life and
// then draws, in printed order, so a "whenever you gain life" payoff
// sees the gain before the cards arrive.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f84d1291-1f82-4b67-a26e-b79624b4ce1d",
		Name:            "Cloudblazer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Cloudblazer — gain 2 life and draw two cards", b33GainLifeThenDraw(2, 2))
			},
		}},
	})
}
