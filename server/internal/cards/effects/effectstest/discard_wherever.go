package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// discard_wherever_test.go — #3859, ADR 0153's 2026-10-15 amendment. A
// card discarded is discarded wherever it goes (CR 701.9a), and "when a
// spell or ability an opponent controls causes you to discard this card"
// looks back at that move (CR 603.10a). So madness (CR 702.35a), Rest in
// Peace and Library of Leng sending the card to exile or a library do not
// stop its own ability. These tests run the ten-card family under Rest in
// Peace (the discard lands in exile), and then the other zones.

const (
	MadnessProbeOracle = "test-discard-madness-probe"
)

var MadnessProbeFired int

func init() {
	// A madness card whose own discard ability counts its firings, so the
	// two ways a discard can end in exile (madness and Rest in Peace) are
	// proved against the real madness machinery and not a stand-in.
	effects.Register(effects.Spec{
		OracleID: MadnessProbeOracle,
		Name:     "Madness Discard Probe",
		Madness:  "{R}",
		Triggered: []game.TriggeredAbility{
			effects.WhereverDiscarded(effects.On(game.EventDiscardCard, effects.DiscardedByAnOpponent(true),
				"Madness Discard Probe — discarded",
				func(*game.Game, *game.StackItem) error { MadnessProbeFired++; return nil })),
		},
	})
}
