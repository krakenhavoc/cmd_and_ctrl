package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sphinx's Disciple — Creature — Human Wizard {3}{U}{U}, 2/2 (Born of the Gods,
// #5559):
//
//	"Flying
//	 Inspired — Whenever this creature becomes untapped, draw a card."
//
// Inspired is an ability word: the trigger is Key to the City's
// (EventUntapCard on itself), so it fires in the untap step and on
// any other untap.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "54634ef2-cb7a-4203-8e48-2733f8c3911b",
		Name:            "Sphinx's Disciple",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventUntapCard, effects.Self, "Sphinx's Disciple — draw a card (inspired)", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
