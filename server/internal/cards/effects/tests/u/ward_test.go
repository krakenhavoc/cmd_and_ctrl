package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Declining the ward payment counters the spell. The giant lives and
// the Doom Blade never resolves.
func TestWardCountersWhenTheCasterDeclines(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := effectstest.PushCatalogPermanent(g, me.ID, "Rimeshield Frost Giant",
		"Creature — Giant Warrior", effectstest.RimeshieldOracle, false)

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := effectstest.CastAtWardedCreature(t, g, opp, giant)

	// The trigger goes on the stack ABOVE the spell and resolves
	// first; its resolution raises the payment prompt — addressed to
	// the CASTER, not to the giant's controller.
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for the spell's controller")
	}
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("the ward permanent's controller must not be asked to pay")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(giant) {
		t.Error("a declined ward must counter the spell — the giant should live")
	}
	if !opp.Graveyard.Contains(blade) {
		t.Error("the countered spell should be in its owner's graveyard")
	}
}

// Paying lets the spell resolve. This is the case a targeting-gate
// implementation could never reach: it would have refused the
// announce and never offered the payment at all.
func TestWardPaidLetsTheSpellResolve(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := effectstest.PushCatalogPermanent(g, me.ID, "Rimeshield Frost Giant",
		"Creature — Giant Warrior", effectstest.RimeshieldOracle, false)

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	// Fund the payment: three Islands in play for the auto-tapper.
	for i := 0; i < 3; i++ {
		g.Battlefield.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Island",
			TypeLine: "Basic Land — Island", Owner: opp.ID, Controller: opp.ID,
		})
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.CastAtWardedCreature(t, g, opp, giant)

	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(giant) {
		t.Error("the ward was paid — the removal spell should have resolved")
	}
}

// "An opponent controls". Your own spell on your own warded creature
// does not trigger it at all — no prompt, no counter.
func TestWardIgnoresItsOwnControllersSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	giant := effectstest.PushCatalogPermanent(g, me.ID, "Rimeshield Frost Giant",
		"Creature — Giant Warrior", effectstest.RimeshieldOracle, false)

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.CastAtWardedCreature(t, g, me, giant)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("ward must not trigger on its own controller's spell")
	}
	if g.Battlefield.Contains(giant) {
		t.Error("your own removal on your own warded creature resolves normally")
	}
}

// Hulking Raptor's other half: "at the beginning of your first main
// phase, add {G}{G}". It rides the S30 EventBeginPrecombatMain
// announcement, and it must fire for its controller's main phase and
// nobody else's.
func TestHulkingRaptorRitualsOnYourPrecombatMain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Hulking Raptor",
		"Creature — Dinosaur", effectstest.HulkingRaptorOracl, false)

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)

	green := 0
	for _, c := range effectstest.Batch01PoolColors(me) {
		if c == "G" {
			green++
		}
	}
	if green != 2 {
		t.Errorf("pool has %d green, want 2", green)
	}
}
