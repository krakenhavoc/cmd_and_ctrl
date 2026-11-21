package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kuja_a_cerberus_test.go — Summon: G.F. Cerberus (#3213).

func castCerberus(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Summon: G.F. Cerberus", TypeLine: "Enchantment Creature — Saga Dog",
		OracleID: effectstest.CerberusOracle, ManaCost: "{2}{R}{R}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell Summon: G.F. Cerberus: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Chapter I: surveil 1 (decline to bin the card).
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID)
	if c == nil {
		t.Fatal("chapter I queued no surveil")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestCerberusChapterIIDoublesTheNextInstantOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, victim := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	cerberus := castCerberus(t, g)
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("chapter I left %d delayed triggers, want 0", len(g.DelayedTriggers))
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat) // chapter II
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("chapter II left %d delayed triggers, want 1", len(g.DelayedTriggers))
	}
	before := effectstest.LifeOf(g, victim.ID)
	effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me.ID, victim.ID, 1)
	if got := effectstest.LifeOf(g, victim.ID); got != before-6 {
		t.Errorf("Double: victim life %d -> %d, want -6 (a Bolt and one copy)", before, got)
	}
	// "Next" instant or sorcery only: a second Bolt is not copied.
	before = effectstest.LifeOf(g, victim.ID)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, victim.ID); got != before-3 {
		t.Errorf("second Bolt: %d -> %d, want -3 (the trigger is spent)", before, got)
	}
	_ = cerberus
}

func TestCerberusChapterIIITriplesTheNextInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, victim := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	cerberus := castCerberus(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat) // II
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat) // III: the earlier delayed trigger expired with its turn
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("chapter III left %d delayed triggers, want 1", len(g.DelayedTriggers))
	}
	before := effectstest.LifeOf(g, victim.ID)
	effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me.ID, victim.ID, 2)
	if got := effectstest.LifeOf(g, victim.ID); got != before-9 {
		t.Errorf("Triple: victim life %d -> %d, want -9 (a Bolt and two copies)", before, got)
	}
	// After chapter III the Saga is sacrificed.
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cerberus) {
		t.Error("the Saga should be sacrificed after chapter III")
	}
}

func TestCerberusIsAThreeThreeCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := castCerberus(t, g)
	if got := effectstest.EffectivePower(t, g, id); got != 3 {
		t.Errorf("power = %d, want 3", got)
	}
}
