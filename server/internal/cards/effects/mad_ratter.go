package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mad Ratter — Creature — Goblin {3}{R}, 1/2 (Throne of Eldraine, #5516):
//
//	"Whenever you draw your second card each turn, create two 1/1 black Rat creature tokens."
//
// Thopter Fabricator's "second card each turn" condition
// (b41DrewYourSecondCardThisTurn): it fires once a turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "779fa24d-e9d2-475d-b3aa-dc0388b18307",
		Name:         "Mad Ratter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Mad Ratter — create two 1/1 black Rat creature tokens",
				Do(CreateToken{Template: TokenCard("1/1 black Rat"), N: 2})),
		},
	})
}
