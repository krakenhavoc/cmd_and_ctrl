package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mystic Skyfish — Creature — Fish {2}{U}, 3/1 (Foundations Jumpstart):
//
//	"Whenever you draw your second card each turn, this creature gains
//	 flying until end of turn."
//
// The draw tally is read after the draw, so the trigger fires on the
// second draw and no other (Alandra's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "645b5784-a6f7-4cf3-966a-e1a51420b96b",
		Name:         "Mystic Skyfish",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Mystic Skyfish — gains flying until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"flying"}, Label: "Mystic Skyfish — flying"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
