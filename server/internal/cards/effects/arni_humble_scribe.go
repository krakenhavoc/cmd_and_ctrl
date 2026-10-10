package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arni, Humble Scribe — Legendary Creature — Human Wizard {2}{U}, 3/2:
//
//	"Whenever another nontoken creature you control enters, untap Arni.
//	 {T}: Draw a card, then discard a card."
//
// The loot discards through the ordinary prompt, opened only after the
// draw has resolved, so the card just drawn is a legal discard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a680d1ab-99f2-43d0-b289-5fdba4083d03",
		Name:         "Arni, Humble Scribe",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				if !AnotherCreatureEnteredUnderYourControl(ev, source, lki, g) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !c.IsToken()
			}, "Arni, Humble Scribe — untap Arni",
				func(g *game.Game, item *game.StackItem) error {
					return UntapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return drawThenDiscard(g, item.Controller, item.SourceCardID, 1, 1, "Arni, Humble Scribe — discard a card")
			},
		}},
	})
}
