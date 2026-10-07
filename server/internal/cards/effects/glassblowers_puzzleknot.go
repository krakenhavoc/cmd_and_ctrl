package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glassblower's Puzzleknot — Artifact {2}:
//
//	"When this artifact enters, scry 2, then you get {E}{E}. (You get two
//	 energy counters. To scry 2, look at the top two cards of your
//	 library, then put any number of them on the bottom and the rest on
//	 top in any order.)
//	 {2}{U}, Sacrifice this artifact: Scry 2, then you get {E}{E}."
//
// ADR 0129 PR 1. The energy follows the scry prompt (scryThenGetEnergy).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0fbe0610-8565-4910-bf35-2c10e38b633a",
		Name:         "Glassblower's Puzzleknot",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Glassblower's Puzzleknot — scry 2, then you get {E}{E}", scryThenGetEnergy(2, 2)),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}, Sacrifice this artifact: Scry 2, then you get {E}{E}.",
			Cost:    Plus(ManaCost("{2}{U}"), SacrificeThis()),
			Purpose: game.Purpose{Energy: 2},
			Effect:  scryThenGetEnergy(2, 2),
		}},
	})
}
