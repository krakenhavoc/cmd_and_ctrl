package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ward_test.go — CR 702.21. The cases that matter are the ones that
// separate ward from hexproof, because a ward implemented at the
// targeting gate would pass none of them:
//
//	announce succeeds        a warded permanent is a LEGAL target
//	the payer is the caster   not the ward permanent's controller
//	paying resolves the spell a refused target could never do this
//	declining counters it     and the removal never resolves
//
// Plus the one hexproof-shaped case ward shares: your own spell on
// your own warded creature does not trigger it.

const (
	rimeshieldOracle   = "ec47cf67-2580-464f-8118-7eabea5be11c"
	hulkingRaptorOracl = "9f2e7533-cfc1-4dd8-a9b8-09cdc712ef2f"
)

// castAtWardedCreature seeds a Doom Blade in `caster`'s hand and
// casts it at `victim`, returning the spell's ID.
func castAtWardedCreature(t *testing.T, g *game.Game, caster *game.Player, victim uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Doom Blade", TypeLine: "Instant",
		OracleID: doomBladeOracle, Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("a warded permanent is a LEGAL target — announce must succeed: %v", err)
	}
	return id
}

// Declining the ward payment counters the spell. The giant lives and
// the Doom Blade never resolves.
func TestWardCountersWhenTheCasterDeclines(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := pushCatalogPermanent(g, me.ID, "Rimeshield Frost Giant",
		"Creature — Giant Warrior", rimeshieldOracle, false)

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := castAtWardedCreature(t, g, opp, giant)

	// The trigger goes on the stack ABOVE the spell and resolves
	// first; its resolution raises the payment prompt — addressed to
	// the CASTER, not to the giant's controller.
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for the spell's controller")
	}
	if hasPayUnlessFor(g, me.ID) {
		t.Error("the ward permanent's controller must not be asked to pay")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := pushCatalogPermanent(g, me.ID, "Rimeshield Frost Giant",
		"Creature — Giant Warrior", rimeshieldOracle, false)

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
	castAtWardedCreature(t, g, opp, giant)

	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	answerPayUnless(t, g, opp.ID, true)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(giant) {
		t.Error("the ward was paid — the removal spell should have resolved")
	}
}

// "An opponent controls". Your own spell on your own warded creature
// does not trigger it at all — no prompt, no counter.
func TestWardIgnoresItsOwnControllersSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	giant := pushCatalogPermanent(g, me.ID, "Rimeshield Frost Giant",
		"Creature — Giant Warrior", rimeshieldOracle, false)

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	castAtWardedCreature(t, g, me, giant)
	passPriorityAroundTable(t, g)

	if hasPayUnlessFor(g, me.ID) {
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Hulking Raptor",
		"Creature — Dinosaur", hulkingRaptorOracl, false)

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)

	green := 0
	for _, c := range batch01PoolColors(me) {
		if c == "G" {
			green++
		}
	}
	if green != 2 {
		t.Errorf("pool has %d green, want 2", green)
	}
}
