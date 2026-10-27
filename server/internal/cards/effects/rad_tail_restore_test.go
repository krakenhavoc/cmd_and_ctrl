package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rad_tail_restore_test.go — #4361 S9: CR 728.1's rad mill and its radiation
// life loss carry the rest of the ability as registered tails
// (engine/rad-milled, engine/rad-life-lost) instead of closures, so a table
// waiting on a CR 616 ordering prompt over either is a restore point.

// radLossDoubleProbeOracle is a test card: "If a player would lose life from
// radiation, they lose twice that much instead." Beside the Life Loss Probe
// (controlled by another player) it makes a radiation loss an ordering
// prompt.
const radLossDoubleProbeOracle = "test-4361-rad-loss-double"

func init() {
	Register(Spec{
		OracleID: radLossDoubleProbeOracle,
		Name:     "Rad Loss Probe",
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventChangeLife},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
				return ev.Kind == game.RepEventLife && ev.LifeFromRadiation && ev.LifeDelta < 0
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.LifeDelta *= 2
				return nil
			},
			Label: "Rad Loss Probe: twice that much",
		}},
	})
}

// radTailTable walks to the active seat's next draw step, gives them a known
// library (top first) and n rad counters, and moves into their precombat main
// phase, where CR 728.1's trigger waits.
func radTailTable(t *testing.T, g *game.Game, n int, library ...string) *game.Player {
	t.Helper()
	seat := g.Turn.ActiveSeat
	advanceToPrecombatMainOf(t, g, seat)
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
	advanceToPrecombatMainOf(t, g, seat)
	return me
}

// The rad mill's amount window, then the radiation loss, each ask for an
// order; a strict restore is taken at each, and the restored table finishes
// CR 728.1: three nonland cards milled, the loss doubled then plus one, and
// three rad counters off.
func TestRadMillAndRadiationLossAreRestorePoints(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	other := g.Seats[(seat+1)%len(g.Seats)]
	seedOrderProbes(g, other.ID, millPlusOneProbeOracles, "Mill Plus One Probe")
	plusOne := seedReplacementPermanent(g, lifeLossPlusOneProbeOracle, "Life Loss Probe", other.ID)
	double := seedReplacementPermanent(g, radLossDoubleProbeOracle, "Rad Loss Probe", other.ID)
	me := radTailTable(t, g, 4, "Bear", "Ogre", "Forest", "Elf", "Forest", "Forest", "Forest", "Forest")
	life, grave := me.Life, me.Graveyard.Size()

	radSettleTurnStart(t, g, me.ID)
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Chooser != me.ID {
		t.Fatalf("pending %+v, want the milling player's ordering prompt", g.PendingChoices)
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMill)
	me = seatByID(g, me.ID)
	if me.Graveyard.Size() != grave {
		t.Fatal("the restored table milled before the answer")
	}
	answerExileOrder(t, g)
	if got := me.Graveyard.Size() - grave; got != 6 {
		t.Fatalf("milled %d after the restored answer, want 6 (four plus one plus one)", got)
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Chooser != me.ID {
		t.Fatalf("pending %+v, want the radiation loss's ordering prompt", g.PendingChoices)
	}
	if me.Life != life {
		t.Fatalf("life moved to %d before the loss's order was answered", me.Life)
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeLifeTail)
	me = seatByID(g, me.ID)
	answerOrderBySource(t, g, double, plusOne)
	if me.Life != life-7 {
		t.Errorf("life %d, want %d: three from radiation, doubled, plus one", me.Life, life-7)
	}
	if got := radOf(me); got != 1 {
		t.Errorf("rad counters %d, want 1: one off for each of the three nonland cards milled", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("questions left open: %+v", g.PendingChoices)
	}
}
