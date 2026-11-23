package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// end_the_turn_cards_test.go — the proof cards for #2165 (CR 724.1, end
// the turn): Sundial of the Infinite, Obeka, Brute Chronologist, Time
// Stop, Glorious End and Ultima. The engine's own tests are
// game/end_turn_test.go.

const (
	SundialOracle     = "79d46e09-2548-440f-9c02-3c3dc39a0cd1"
	ObekaOracle       = "21b76a94-d9f3-4c34-ab24-7e89321b04f0"
	TimeStopOracle    = "55bc6a7e-2356-48f0-a93a-8bf19044ee4b"
	GloriousEndOracle = "e611a3e0-eb0e-466e-a771-51310eeb34cd"
	UltimaOracle      = "c9a45c28-826e-4e6e-8adc-1368ea1af11a"
	EndTurnTestLabel  = "end-the-turn test — exile it at the next end step"
)

// discardIfOwed answers the cleanup step's hand-size discard (CR 514.1)
// in the turn that ended, with the first cards in hand. newCatalogGame's
// active seat holds eight cards after its draw, so the ended turn's
// cleanup step asks for one before the next turn can begin.
func DiscardIfOwed(t *testing.T, g *game.Game) {
	t.Helper()
	if g.Turn.Step != game.StepCleanup {
		return
	}
	for _, p := range g.Seats {
		n := g.DiscardPending[p.ID]
		if n <= 0 {
			continue
		}
		ids := make([]uuid.UUID, 0, n)
		for i := 0; i < n && i < len(p.Hand.Cards); i++ {
			ids = append(ids, p.Hand.Cards[i].InstanceID)
		}
		if err := g.DiscardSelection(p.ID, ids); err != nil {
			t.Fatalf("DiscardSelection: %v", err)
		}
	}
}
