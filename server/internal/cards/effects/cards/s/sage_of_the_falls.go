package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sage of the Falls — Creature — Merfolk Wizard {4}{U}, 2/5 (Jumpstart 2022, #5948):
//
//	"Whenever this creature or another non-Human creature you control
//	 enters, you may draw a card. If you do, discard a card."
//
// One trigger on every creature that enters under your control (this one
// included) whose creature types, read through the layers as it enters,
// don't include Human. "Human" is a word of the Sage's text (CR 612.2).
// The loot is Stadium Tidalmage's "may draw, if you do discard".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "852621bf-9ab7-4027-89fb-b438e4080684",
		Name:         "Sage of the Falls",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && !c.HasSubtype(source.TextSubtype("Human"))
			}, "Sage of the Falls — you may draw a card; if you do, discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Sage of the Falls — draw a card, then discard a card?",
						Draws:    1,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     effects.CelestusAnswer,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
