package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ADR 0167 §2, CR 723.1b: "If the targeted player skips their next
// turn, you'll control the next turn the affected player actually
// takes, and the extra turn the player takes will be after that turn."
// The controlled player's next turn is an extra turn Trouble in Pairs
// skips; the control waits for the turn after, which they do take.
func TestTurnControlWaitsPastASkippedTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	n := len(g.Seats)
	opp := (me + 2) % n
	effectstest.SeedEnchantment(g, effectstest.TroubleInPairsOracle, "Trouble in Pairs", "Enchantment", g.Seats[me].ID)
	effectstest.CastCatalogSpell(t, g, "Time Warp", "Sorcery", effectstest.TimeWarpOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[opp].ID}})
	effectstest.PassPriorityAroundTable(t, g)
	var ok bool
	g.WithWriteLock(func() {
		ok = g.BeginTurnControlForEffect(game.TurnControl{
			Controller: g.Seats[me].ID, Controlled: g.Seats[opp].ID,
			Source: uuid.New(), SourceName: "Emrakul, the Promised End",
		})
	})
	if !ok {
		t.Fatal("BeginTurnControlForEffect refused")
	}

	effectstest.EndTurn(t, g)
	// Time Warp's turn was skipped: the next seat's normal turn.
	effectstest.AssertTurnOf(t, g, (me+1)%n, false, "after the skipped extra turn")
	if got := g.ControllerOfPlayer(g.Seats[opp].ID); got != g.Seats[opp].ID {
		t.Fatal("the control applied without its turn")
	}
	var waiting int
	g.ReadSnapshot(func() { waiting = len(g.ScheduledPlayerControlsLocked()) })
	if waiting != 1 {
		t.Fatalf("%d controls waiting after the skip, want 1", waiting)
	}

	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, opp, false, "the next turn the player actually takes")
	if got := g.ControllerOfPlayer(g.Seats[opp].ID); got != g.Seats[me].ID {
		t.Fatalf("the turn they took is controlled by %v, want the controller", got)
	}
	effectstest.EndTurn(t, g)
	// The extra turn is theirs, and Trouble in Pairs skips it too: the
	// promise is kept by queuing it, and the skip is the enchantment's.
	effectstest.AssertTurnOf(t, g, (opp+1)%n, false, "after the controlled turn and its skipped extra turn")
	skipped := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventExtraTurnSkipped && ev.Actor == g.Seats[opp].ID {
			skipped++
		}
	}
	if skipped != 2 {
		t.Errorf("%d of the opponent's extra turns skipped, want 2 (Time Warp's and the control's)", skipped)
	}
}
