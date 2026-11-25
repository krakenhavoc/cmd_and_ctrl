package testscr

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// devour_x_test.go — "Devour X, where X is the number of creatures
// devoured this way" (Thromok the Insatiable, #6267), on a test oracle.
//
//	702.82a "Devour N" means "As this object enters, you may sacrifice
//	        any number of creatures. This permanent enters with N +1/+1
//	        counters on it for each creature sacrificed this way."
//	702.82b "It devoured" means "sacrificed as a result of its devour
//	        ability as it entered the battlefield."

const testDevourXOracle = "test-6267-devour-x"

func registerDevourX(t *testing.T) {
	t.Helper()
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:     testDevourXOracle,
		Name:         "Test Devour X",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.DevourX("Test Devour X")},
	})
}

// k creatures devoured: X = k, so k*k counters, and Card.Devoured = k.
func TestDevourXPutsXCountersForEachCreatureDevoured(t *testing.T) {
	for _, k := range []int{0, 1, 3} {
		t.Run(fmt.Sprintf("devour %d", k), func(t *testing.T) {
			registerDevourX(t)
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			var fodder []uuid.UUID
			for i := 0; i < 3; i++ {
				fodder = append(fodder, effectstest.PushDevourFodder(g, me.ID, "Goblin"))
			}
			id := effectstest.CastDevourer(t, g, "Test Devour X", testDevourXOracle)
			for _, c := range protocol.ViewOfGameFor(g, me.ID.String()).PendingChoices {
				if c.Kind == string(game.PendingChoiceEntrySacrifice) && (!c.DevourX || c.Devour != 0) {
					t.Fatalf("k=%d: the prompt reads devour_x=%v devour=%d, want devour_x and no fixed N", k, c.DevourX, c.Devour)
				}
			}
			effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, fodder[:k]...)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.DevourPlusOnes(g, id); got != k*k {
				t.Fatalf("devoured %d: +1/+1 counters = %d, want %d", k, got, k*k)
			}
			if got := g.DevouredBy(id); got != k {
				t.Fatalf("devoured %d: DevouredBy = %d", k, got)
			}
		})
	}
}
