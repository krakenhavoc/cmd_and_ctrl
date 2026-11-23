package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// extra_turn_cards_test.go — the CR 500.7 first wave (ADR 0059 "Card
// first wave", #753). Each card is driven through the real cast or
// activation and the real rotation seam: the extra turn is asserted by
// who is active after the current turn ends, never by reading the
// queue alone.

const (
	TimeWarpOracle             = "dbd6a94b-62ff-4a10-9d52-bdd90b26e425"
	TimeStretchOracle          = "72e56963-a9dd-44dc-a4d3-992b4d89dd28"
	TemporalManipulationOracle = "6b6ac99a-7548-4cad-9fe5-ea611618ab9e"
	CaptureOfJingzhouOracle    = "89aa65d9-2502-40b0-90b6-b25a8e9f6155"
	TemporalMasteryOracle      = "5c58b8e6-c572-461e-893e-a8c05f20ba17"
	FinalFortuneOracle         = "8d0adc5c-3fdd-4e22-b783-5651f8e57b65"
	LastChanceOracle           = "360039a5-1cbd-4ee3-8f94-21b5348e106a"
	WarriorsOathOracle         = "574044cf-2e2f-4b5c-b4d1-cf05ba814ab2"
	MagistratesScepterOracle   = "487485a6-cf79-48c0-bea9-0ec6b3ee253c"
)

// endTurn ends the current turn through the rotation seam. The sandbox
// pass_turn verb is used rather than walking cleanup, because the
// catalog harness's starting hands are over the hand size and a real
// cleanup would stop on the discard.
func EndTurn(t *testing.T, g *game.Game) {
	t.Helper()
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
}

// assertTurnOf fails unless it is `seat`'s turn and its extra-ness is `extra`.
func AssertTurnOf(t *testing.T, g *game.Game, seat int, extra bool, when string) {
	t.Helper()
	if g.Turn.ActiveSeat != seat || g.Turn.Extra != extra {
		t.Fatalf("%s: seat %d extra=%v, want seat %d extra=%v", when, g.Turn.ActiveSeat, g.Turn.Extra, seat, extra)
	}
}
