package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Weldfast Wingsmith — Creature — Human Artificer {3}{U}, 3/3 (Jumpstart 2022, #5948):
//
//	"Whenever an artifact you control enters, this creature gains flying
//	 until end of turn."
//
// Captain Storm's "artifact entered under your control" watcher with this
// creature's own until-end-of-turn keyword grant. The Wingsmith is not an
// artifact, so it doesn't trigger on itself; an artifact creature entering
// does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a05bcc3e-25d5-4111-820e-a097189afb7e",
		Name:         "Weldfast Wingsmith",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.IsArtifactEnteredUnderYourControl(ev, source, g)
			}, "Weldfast Wingsmith — this creature gains flying until end of turn",
				effects.ThisGainsKeywordUntilEndOfTurn("flying", "Weldfast Wingsmith — flying until end of turn")),
		},
	})
}
