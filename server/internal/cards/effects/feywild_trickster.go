package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feywild Trickster — Creature — Gnome Warlock {2}{U}, 2/2 (Adventures in
// the Forgotten Realms):
//
//	"Whenever you roll one or more dice, create a 1/1 blue Faerie Dragon
//	 creature token with flying."
//
// One trigger per instruction to roll, however many dice (WheneverYouRollDice).
func init() {
	Register(Spec{
		OracleID: "b385edc3-c897-48e3-b251-242c77159fa0", Name: "Feywild Trickster", Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouRollDice("Feywild Trickster — create a 1/1 blue Faerie Dragon with flying",
				func(g *game.Game, item *game.StackItem, _ int) error {
					return CreateToken{Template: TokenCard("1/1 blue Faerie Dragon with flying"), N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
