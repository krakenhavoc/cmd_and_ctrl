package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kuja_a_cerberus_test.go — Summon: G.F. Cerberus (#3213).

func castCerberus(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Summon: G.F. Cerberus", TypeLine: "Enchantment Creature — Saga Dog",
		OracleID: cerberusOracle, ManaCost: "{2}{R}{R}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell Summon: G.F. Cerberus: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Chapter I: surveil 1 (decline to bin the card).
	c := latestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID)
	if c == nil {
		t.Fatal("chapter I queued no surveil")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func TestCerberusChapterIIDoublesTheNextInstantOnly(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, victim := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	cerberus := castCerberus(t, g)
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("chapter I left %d delayed triggers, want 0", len(g.DelayedTriggers))
	}

	advanceToPrecombatMainOf(t, g, seat) // chapter II
	passPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("chapter II left %d delayed triggers, want 1", len(g.DelayedTriggers))
	}
	before := lifeOf(g, victim.ID)
	castBoltAtAndKeepTheCopysTarget(t, g, me.ID, victim.ID, 1)
	if got := lifeOf(g, victim.ID); got != before-6 {
		t.Errorf("Double: victim life %d -> %d, want -6 (a Bolt and one copy)", before, got)
	}
	// "Next" instant or sorcery only: a second Bolt is not copied.
	before = lifeOf(g, victim.ID)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, victim.ID); got != before-3 {
		t.Errorf("second Bolt: %d -> %d, want -3 (the trigger is spent)", before, got)
	}
	_ = cerberus
}

func TestCerberusChapterIIITriplesTheNextInstant(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, victim := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	cerberus := castCerberus(t, g)
	advanceToPrecombatMainOf(t, g, seat) // II
	passPriorityAroundTable(t, g)
	advanceToPrecombatMainOf(t, g, seat) // III: the earlier delayed trigger expired with its turn
	passPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("chapter III left %d delayed triggers, want 1", len(g.DelayedTriggers))
	}
	before := lifeOf(g, victim.ID)
	castBoltAtAndKeepTheCopysTarget(t, g, me.ID, victim.ID, 2)
	if got := lifeOf(g, victim.ID); got != before-9 {
		t.Errorf("Triple: victim life %d -> %d, want -9 (a Bolt and two copies)", before, got)
	}
	// After chapter III the Saga is sacrificed.
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cerberus) {
		t.Error("the Saga should be sacrificed after chapter III")
	}
}

func TestCerberusIsAThreeThreeCreature(t *testing.T) {
	g := newCatalogGame(t)
	id := castCerberus(t, g)
	if got := effectivePower(t, g, id); got != 3 {
		t.Errorf("power = %d, want 3", got)
	}
}
