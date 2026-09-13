package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Consecrated Sphinx — 4/6 Creature — Sphinx for {4}{U}{U}:
//
//	"Flying. Whenever an opponent draws a card, you may draw two
//	cards."
//
// S19 sub-PR 6: an optional opponent-draws trigger. EventDrawCard
// fires per card, so an opponent's "draw three" queues three
// yes/no prompts for the Sphinx's controller — paper behaviour.
// Only opponents' draws count: the Sphinx's own two-card draws
// are the controller's and never re-trigger it (no loop).
// OptionalPrompt puts the "you may" in the controller's hands;
// "Yes" puts the trigger on the stack and the draw happens on
// resolution.
func init() {
	Register(Spec{
		OracleID:        "311a449d-dc74-46e6-9a47-6a597931f736",
		Name:            "Consecrated Sphinx",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDrawCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil && ev.Actor != source.Controller
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Consecrated Sphinx — draw two cards",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
					})
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Consecrated Sphinx — draw two cards?",
			},
		}},
	})
}
