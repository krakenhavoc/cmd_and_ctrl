package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const persistentMarshstalkerOracle = "77eb76cf-a361-44e0-aeee-e8ef4b695799"

func TestPersistentMarshstalkerIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(persistentMarshstalkerOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Persistent Marshstalker: %+v %v", spec, ok)
	}
}

// pmGraveyard seeds the Marshstalker and `filler` other cards into me's
// graveyard (so the graveyard holds filler+1 cards).
func pmGraveyard(me *game.Player, filler int) uuid.UUID {
	for i := 0; i < filler; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "Persistent Marshstalker", TypeLine: "Creature — Rat Berserker", OracleID: persistentMarshstalkerOracle,
		ManaCost: "{1}{B}", Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	return id
}

func TestPersistentMarshstalkerGetsPlusOnePowerPerOtherRat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pm := effectstest.DskB1Creature(g, me.ID, "Persistent Marshstalker", "Creature — Rat Berserker", persistentMarshstalkerOracle, 3, 1)
	effectstest.B12Creature(g, me.ID, "Rat A", "Creature — Rat", 1, 1)
	effectstest.B12Creature(g, me.ID, "Rat B", "Creature — Rat", 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	g.RunStateChecksForTest()
	c, _ := effectstest.BattlefieldCard(g, pm)
	if e := c.Effective(); e.Power != 5 || e.Toughness != 1 {
		t.Errorf("Marshstalker %d/%d with two other Rats, want 5/1", e.Power, e.Toughness)
	}
}

func TestPersistentMarshstalkerReturnsTappedAndAttackingWithThreshold(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rat := effectstest.PushCatalogPermanent(g, me.ID, "Rat", "Creature — Rat", "", false)
	pm := pmGraveyard(me, 6) // seven cards
	effectstest.DeclareAttack(t, g, opp.ID, rat)
	if n := effectstest.TriggersOnStackFrom(g, pm); n != 1 {
		t.Fatalf("triggers = %d, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{B}{B}{B}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if pick := effectstest.LatestOptionPickFor(g, me.ID); pick != nil {
		effectstest.AnswerOptionPick(t, g, me.ID, 0)
	}
	c, ok := effectstest.InterceptorOnBattlefield(g, pm)
	if !ok {
		t.Fatal("the Marshstalker did not return")
	}
	if !c.Tapped || c.AttackingTarget == uuid.Nil {
		t.Errorf("tapped=%v attacking=%v, want tapped and attacking", c.Tapped, c.AttackingTarget)
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventAttack && ev.CardID == pm {
			t.Error("the returned Marshstalker was declared as an attacker")
		}
	}
}

func TestPersistentMarshstalkerNeedsSevenCardsAndARat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rat := effectstest.PushCatalogPermanent(g, me.ID, "Rat", "Creature — Rat", "", false)
	pm := pmGraveyard(me, 5) // six cards
	effectstest.DeclareAttack(t, g, opp.ID, rat)
	if n := effectstest.TriggersOnStackFrom(g, pm); n != 0 {
		t.Fatalf("six cards in the graveyard triggered %d times", n)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	bear := effectstest.PushCatalogPermanent(g2, me2.ID, "Bear", "Creature — Bear", "", false)
	pm2 := pmGraveyard(me2, 6)
	effectstest.DeclareAttack(t, g2, opp2.ID, bear)
	if n := effectstest.TriggersOnStackFrom(g2, pm2); n != 0 {
		t.Fatalf("a Bear attack triggered %d times", n)
	}
}

func TestPersistentMarshstalkerDeclineLeavesItInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rat := effectstest.PushCatalogPermanent(g, me.ID, "Rat", "Creature — Rat", "", false)
	pm := pmGraveyard(me, 6)
	effectstest.DeclareAttack(t, g, opp.ID, rat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if !me.Graveyard.Contains(pm) {
		t.Error("declining moved the Marshstalker")
	}
}
