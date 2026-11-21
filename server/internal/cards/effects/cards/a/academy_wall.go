package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Academy Wall — Creature — Wall {2}{U}, 0/5 (Dominaria United, #5487):
//
//	"Defender
//	 Whenever you cast an instant or sorcery spell, you may draw a card. If
//	 you do, discard a card. This ability triggers only once each turn."
//
// The loot is Daring Saboteur's "you may draw, then discard" prompt on a
// cast trigger held to once each turn per object (Harnesser of Storms'
// TriggersOnlyOnceEachTurn): a second spell the same turn does not even
// trigger it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0cd77ee1-408c-4c51-ac87-24f4fc50083d",
		Name:            "Academy Wall",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.TriggersOnlyOnceEachTurn(effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
				"Academy Wall — you may draw a card; if you do, discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Academy Wall — draw a card, then discard a card?",
						Draws:    1,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     effects.CelestusAnswer,
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
