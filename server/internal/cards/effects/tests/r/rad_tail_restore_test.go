package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// radTailTable walks to the active seat's next draw step, gives them a known
// library (top first) and n rad counters, and moves into their precombat main
// phase, where CR 728.1's trigger waits.
func radTailTable(t *testing.T, g *game.Game, n int, library ...string) *game.Player {
	t.Helper()
	seat := g.Turn.ActiveSeat
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	for i := 0; i < 400 && (g.Turn.Step != game.StepDraw || g.Turn.ActiveSeat != seat); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward the draw step: %v", err)
		}
	}
	me := g.Seats[seat]
	g.WithWriteLock(func() {
		me.Library.Cards = nil
		for i := len(library) - 1; i >= 0; i-- {
			c := game.Card{InstanceID: uuid.New(), Name: library[i], Owner: me.ID, Controller: me.ID, TypeLine: "Creature — Bear"}
			if library[i] == "Forest" {
				c.TypeLine = "Basic Land — Forest"
			}
			me.Library.PushTop(c)
		}
	})
	radGive(t, g, me, n)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	return me
}

// The rad mill's amount window, then the radiation loss, each ask for an
// order; a strict restore is taken at each, and the restored table finishes
// CR 728.1: three nonland cards milled, the loss doubled then plus one, and
// three rad counters off.
func TestRadMillAndRadiationLossAreRestorePoints(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	other := g.Seats[(seat+1)%len(g.Seats)]
	effectstest.SeedOrderProbes(g, other.ID, effectstest.MillPlusOneProbeOracles, "Mill Plus One Probe")
	plusOne := effectstest.SeedReplacementPermanent(g, effectstest.LifeLossPlusOneProbeOracle, "Life Loss Probe", other.ID)
	double := effectstest.SeedReplacementPermanent(g, effectstest.RadLossDoubleProbeOracle, "Rad Loss Probe", other.ID)
	me := radTailTable(t, g, 4, "Bear", "Ogre", "Forest", "Elf", "Forest", "Forest", "Forest", "Forest")
	life, grave := me.Life, me.Graveyard.Size()

	radSettleTurnStart(t, g, me.ID)
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Chooser != me.ID {
		t.Fatalf("pending %+v, want the milling player's ordering prompt", g.PendingChoices)
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMill)
	me = effectstest.SeatByID(g, me.ID)
	if me.Graveyard.Size() != grave {
		t.Fatal("the restored table milled before the answer")
	}
	effectstest.AnswerExileOrder(t, g)
	if got := me.Graveyard.Size() - grave; got != 6 {
		t.Fatalf("milled %d after the restored answer, want 6 (four plus one plus one)", got)
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Chooser != me.ID {
		t.Fatalf("pending %+v, want the radiation loss's ordering prompt", g.PendingChoices)
	}
	if me.Life != life {
		t.Fatalf("life moved to %d before the loss's order was answered", me.Life)
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeLifeTail)
	me = effectstest.SeatByID(g, me.ID)
	effectstest.AnswerOrderBySource(t, g, double, plusOne)
	if me.Life != life-7 {
		t.Errorf("life %d, want %d: three from radiation, doubled, plus one", me.Life, life-7)
	}
	if got := effectstest.RadOf(me); got != 1 {
		t.Errorf("rad counters %d, want 1: one off for each of the three nonland cards milled", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("questions left open: %+v", g.PendingChoices)
	}
}
