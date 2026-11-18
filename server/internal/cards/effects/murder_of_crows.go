package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Murder of Crows — {3}{U}{U} Creature — Bird 4/4:
//
//	"Flying
//	 Whenever another creature dies, you may draw a card. If you do,
//	 discard a card."
//
// Deathbringer Thoctar's "another creature dies" watcher (anyone's
// creature, the source's own death excluded by ID) with a "you may"
// prompt on the trigger. Saying yes is the loot: draw, then discard,
// through the shared draw-then-discard primitive, so the discard is
// owed only after the draw happened ("if you do").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "babab2f5-fd1f-4e0a-a692-6c2ce35df708",
		Name:            "Murder of Crows",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Draws: 1, Answers: game.AnswerValue},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b15AnotherCreatureDied(ev, source, g)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Murder of Crows — draw a card, then discard a card?"},
			Key:            "Murder of Crows — you may draw a card; if you do, discard a card",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return drawThenDiscard(g, item.Controller, item.SourceCardID, 1, 1, "")
			},
		}},
	})
}
