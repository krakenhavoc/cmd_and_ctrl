package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Salvage Drone — Creature — Eldrazi Drone {U}, 1/1 (Battle for Zendikar,
// #5547):
//
//	"Devoid (This card has no color.)
//	 Ingest (Whenever this creature deals combat damage to a player, that
//	 player exiles the top card of their library.)
//	 When this creature dies, you may draw a card. If you do, discard a card."
//
// Ingest is OldB307Ingest. The dies trigger is The Celestus's optional loot
// (the keyed "draw, then discard" answer): drawing is the "may", and the
// discard follows only if you chose to draw.
//
// No simplification.
func init() {
	const name = "Salvage Drone"
	effects.Register(effects.Spec{
		OracleID:        "9b18bbf3-3b1c-4010-bc6f-e7eacb2ffb70",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			effects.OldB307Ingest(name),
			effects.WhenThisDies(name+" — you may draw a card. If you do, discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: name + " — draw a card, then discard a card?",
						Draws:    1,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     effects.CelestusAnswer,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
