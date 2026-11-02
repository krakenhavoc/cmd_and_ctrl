package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spider-Girl, Legacy Hero — Legendary Creature — Spider Human Hero {G}{W}, 2/2
// (Marvel's Spider-Man, #4460):
//
//	"During your turn, Spider-Girl has flying.
//	 When Spider-Girl leaves the battlefield, create a 1/1 green and white Human
//	 Citizen creature token."
//
// Colossus, Steel Stalwart's "during your turn, <self> has <keyword>" layer 6
// grant, and City Pigeon's leaves-the-battlefield trigger with a token.
//
// No simplification.
func init() {
	flying := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return selfOnly(target, g, source) && IsYourTurn(g, source.Controller)
	}, "flying")
	flying.Label = "During your turn, Spider-Girl has flying."
	Register(Spec{
		OracleID:     "d459b3f5-a18c-4c17-a855-dc147f7c2161",
		Name:         "Spider-Girl, Legacy Hero",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{flying},
		Triggered: []game.TriggeredAbility{
			WhenThisLeaves("Spider-Girl — create a 1/1 green and white Human Citizen creature token",
				Do(CreateToken{Template: TokenCard("1/1 green and white Human Citizen"), N: 1})),
		},
	})
}
