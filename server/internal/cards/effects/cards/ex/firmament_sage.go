package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Firmament Sage — {3}{U} Creature — Human Wizard 2/3 (#2561, ADR 0132):
//
//	"If it's neither day nor night, it becomes day as this creature enters.
//	 Whenever day becomes night or night becomes day, draw a card."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57163ade-fe89-4ba0-98a9-daeeed5badb9",
		Name:         "Firmament Sage",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.BecomesDayAsEnters(),
		Triggered: []game.TriggeredAbility{
			effects.WheneverDayBecomesNightOrNightBecomesDay("Firmament Sage — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
