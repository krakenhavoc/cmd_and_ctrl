package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stormfist Crusader — Creature — Human Knight {B}{R}, 2/2 (EDHREC
// rank 1942):
//
//	"Menace
//	 At the beginning of your upkeep, each player draws a card and
//	 loses 1 life."
//
// The symmetrical Rakdos card engine. A "your upkeep" trigger that
// walks the table APNAP from the active seat: every player draws
// (so every "whenever an opponent draws" watcher fires) and then
// loses 1 — b18EachPlayerDrawsAndLosesLife.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0d3bebc6-662a-4cd4-ad1e-aeed0c5e04f6",
		Name:            "Stormfist Crusader",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Stormfist Crusader — each player draws a card and loses 1 life", effects.B18EachPlayerDrawsAndLosesLife),
		},
	})
}
