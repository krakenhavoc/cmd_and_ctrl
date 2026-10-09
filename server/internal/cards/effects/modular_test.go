package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// modular_test.go — #2012, CR 702.43. Modular N is a numbered keyword
// token the engine derives two abilities from: an entry replacement per
// instance ("enters with N +1/+1 counters") and a dies trigger per
// instance ("you may put a +1/+1 counter on target artifact creature
// for each +1/+1 counter on this permanent"). game/modular.go.

func TestModularIsANumberedCumulativeToken(t *testing.T) {
	if toks, ok := game.CanonicalKeywords("Modular 2"); !ok || len(toks) != 1 || toks[0] != "modular 2" {
		t.Errorf("CanonicalKeywords(\"Modular 2\") = %v, %v", toks, ok)
	}
	if _, ok := game.CanonicalKeywords("Modular"); ok {
		t.Error("a bare \"Modular\" was accepted; the number lives on the line")
	}
	if !game.KeywordIsCumulative("modular 1") {
		t.Error("modular is not cumulative; CR 702.43b")
	}
}

// arcboundInHand is a 0/0 artifact creature with the given keyword
// tokens and no catalog entry, in the active player's hand.
func arcboundInHand(g *game.Game, name string, keywords ...string) uuid.UUID {
	me := g.Seats[g.Turn.ActiveSeat]
	c := game.NewCard(name, me.ID)
	c.TypeLine = "Artifact Creature — Construct"
	c.Keywords = keywords
	me.Hand.PushTop(c)
	return c.InstanceID
}

func castArcbound(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// "Enters with N +1/+1 counters": a 0/0 survives, and two instances
// each work (CR 702.43b).
func TestModularEntersWithItsCounters(t *testing.T) {
	g := newCatalogGame(t)
	worker := arcboundInHand(g, "Arcbound Worker", "modular 1")
	castArcbound(t, g, worker)
	if !g.Battlefield.Contains(worker) || countersOn(g, worker, game.CounterPlusOne) != 1 {
		t.Fatalf("modular 1: on battlefield %v with %d counters, want true and 1",
			g.Battlefield.Contains(worker), countersOn(g, worker, game.CounterPlusOne))
	}
	double := arcboundInHand(g, "Double Arcbound", "modular 1", "modular 2")
	castArcbound(t, g, double)
	if got := countersOn(g, double, game.CounterPlusOne); got != 3 {
		t.Errorf("modular 1 and modular 2 gave %d counters, want 3", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("two modular instances asked a question: %+v", g.PendingChoices[0])
	}
}

// Every entry, not only a cast: a modular creature returned from a
// graveyard enters with its counters too.
func TestModularCountsOnAReanimation(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	c := game.NewCard("Arcbound Prototype", me.ID)
	c.TypeLine = "Artifact Creature — Assembly-Worker"
	c.Keywords = []string{"modular 2"}
	me.Graveyard.PushTop(c)
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardForEffect(c.InstanceID, game.ZoneBattlefield); err != nil {
			t.Fatalf("return: %v", err)
		}
	})
	if got := countersOn(g, c.InstanceID, game.CounterPlusOne); got != 2 {
		t.Errorf("reanimated modular 2 has %d counters, want 2", got)
	}
}

// Arcbound Slasher, modular 4 and riot: modular is applied without an
// ordering question, and riot still asks its own.
func TestModularBesideRiotAsksOnlyRiot(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	slasher := arcboundInHand(g, "Arcbound Slasher", "modular 4", "riot")
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, slasher, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceEntryRiot {
		var kinds []game.PendingChoiceKind
		for _, c := range g.PendingChoices {
			kinds = append(kinds, c.Kind)
		}
		t.Fatalf("open prompts %v, want only riot's", kinds)
	}
	if err := g.ResolveEntryRiot(g.PendingChoices[0].ID, me.ID, true); err != nil {
		t.Fatalf("answer riot: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, slasher, game.CounterPlusOne); got != 5 {
		t.Errorf("modular 4 and riot's counter gave %d, want 5", got)
	}
}

// The dies trigger: a "you may" with a target artifact creature, one
// counter per +1/+1 counter the modular permanent last had.
func TestModularMovesItsCountersWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	worker := pushDiesKeywordCreature(g, me.ID, "Arcbound Worker", 0, 0, "modular 1")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == worker {
				g.Battlefield.Cards[i].TypeLine = "Artifact Creature — Construct"
			}
		}
	})
	addCounters(t, g, worker, game.CounterPlusOne, 3)
	artifact := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ornithopter", TypeLine: "Artifact Creature — Thopter",
		Power: 0, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	bear := pushDiesKeywordCreature(g, me.ID, "Bear", 2, 2)

	b25Destroy(g, worker)
	if latestTriggerPrompt(g, me.ID) == nil {
		t.Fatal("no \"you may\" for the modular trigger")
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the modular trigger")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Error("a non-artifact creature was accepted as the target")
	}
	answerPickTarget(t, g, artifact)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, artifact, game.CounterPlusOne); got != 3 {
		t.Errorf("the artifact creature got %d counters, want 3", got)
	}
}

// Arcbound Ravager: enters with its modular counter, grows by
// sacrificing an artifact, and can sacrifice itself to move its
// counters onto another artifact creature.
func TestArcboundRavagerEatsItselfAndPassesItsCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	ravager := game.NewCard("Arcbound Ravager", me.ID)
	ravager.TypeLine, ravager.OracleID = "Artifact Creature — Beast", "62e7e7b1-9887-4d15-b0e5-a8ddc711bd88"
	me.Hand.PushTop(ravager)
	if err := g.CastSpell(me.ID, ravager.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, ravager.InstanceID, game.CounterPlusOne); got != 1 {
		t.Fatalf("the Ravager entered with %d counters, want 1 (modular 1)", got)
	}
	fodder := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ichor Wellspring", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID,
	})
	thopter := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ornithopter", TypeLine: "Artifact Creature — Thopter",
		Power: 0, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	sac := func(id uuid.UUID) {
		t.Helper()
		if err := g.ActivateCatalogAbility(me.ID, ravager.InstanceID, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{id}}); err != nil {
			t.Fatalf("activate: %v", err)
		}
	}
	sac(fodder)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, ravager.InstanceID, game.CounterPlusOne); got != 2 {
		t.Fatalf("after eating an artifact the Ravager has %d counters, want 2", got)
	}
	sac(ravager.InstanceID)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	answerPickTarget(t, g, thopter)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, thopter, game.CounterPlusOne); got != 2 {
		t.Errorf("the Thopter got %d counters from the Ravager's modular, want 2", got)
	}
}
