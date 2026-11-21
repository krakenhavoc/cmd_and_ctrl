package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inga Rune-Eyes — Legendary Creature — Human Wizard {3}{U}, 3/3:
//
//	"When Inga Rune-Eyes enters, scry 3.
//	 When Inga Rune-Eyes dies, draw three cards if three or more creatures
//	 died this turn."
//
// The death clause is a resolution-time condition (it is not an
// intervening "if"), counted from the turn's creature-death tally, which
// already holds Inga herself and counts tokens and last-known creatures.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78afa295-c9f4-4a81-9f7b-5746a3cf6b31",
		Name:         "Inga Rune-Eyes",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 3},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Inga Rune-Eyes — scry 3", effects.Do(effects.Scry{N: 3})),
			effects.WhenThisDies("Inga Rune-Eyes — draw three cards if three or more creatures died this turn",
				func(g *game.Game, item *game.StackItem) error {
					if effects.B11CreaturesDiedThisTurn(g) < 3 {
						return nil
					}
					return effects.DrawCards{Player: item.Controller, N: 3}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
