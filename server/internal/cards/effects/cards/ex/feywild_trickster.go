package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feywild Trickster — Creature — Gnome Warlock {2}{U}, 2/2 (Adventures in
// the Forgotten Realms):
//
//	"Whenever you roll one or more dice, create a 1/1 blue Faerie Dragon
//	 creature token with flying."
//
// One trigger per instruction to roll, however many dice (WheneverYouRollDice).
func init() {
	effects.Register(effects.Spec{
		OracleID: "b385edc3-c897-48e3-b251-242c77159fa0", Name: "Feywild Trickster", Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouRollDice("Feywild Trickster — create a 1/1 blue Faerie Dragon with flying",
				func(g *game.Game, item *game.StackItem, _ int) error {
					return effects.CreateToken{Template: effects.TokenCard("1/1 blue Faerie Dragon with flying"), N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
