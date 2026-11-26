package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b519_test.go — From the Vault: Dragons, enters and dies triggers and
// artifacts: S82 OLD batch 519 (#5589), each card driven through the engine.

const (
	b519EbonDragon = "12e4d2bd-83e9-4120-a2e8-0645c0ed2387"
	b519Draco      = "6157ade6-9037-4339-869b-493fdee71dcf"
)

func TestS82OldB519EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{b519EbonDragon: "Ebon Dragon", b519Draco: "Draco"} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82OldB519EbonDragonMayHaveAnOpponentDiscardACardOfTheirChoice(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		for opp.Hand.Size() > 0 {
			_, _ = opp.Hand.PopTop()
		}
		keep := effectstest.EoHand(opp, "Keep", "Sorcery")
		toss := effectstest.EoHand(opp, "Toss", "Sorcery")
		dragon := effectstest.CastCatalogSpell(t, g, "Ebon Dragon", "Creature — Dragon", b519EbonDragon, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			effectstest.PickPlayer(t, g, me.ID, opp.ID)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.B18TriggerPromptCount(g, me.ID) > 0 {
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		}
		if yes {
			// The opponent picks the card; it is not random.
			effectstest.AnswerDiscard(t, g, opp.ID, toss)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(dragon) || !effectstest.HasKeywordOnBattlefield(t, g, dragon, "flying") {
			t.Fatalf("yes=%v: the Dragon is not a flier on the battlefield", yes)
		}
		if yes && (opp.Hand.Contains(toss) || !opp.Hand.Contains(keep)) {
			t.Errorf("yes: the opponent did not discard exactly the card they chose")
		}
		if !yes && opp.Hand.Size() != 2 {
			t.Errorf("no: the opponent's hand is %d, want 2", opp.Hand.Size())
		}
	}
}

func TestS82OldB519DracoCostsTwoLessPerBasicLandType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const tl = "Artifact Creature — Dragon"
	if got := effectstest.SelfPricedMV(t, g, me, b519Draco, tl, "{16}"); got != 16 {
		t.Fatalf("no lands: %d, want 16", got)
	}
	effectstest.DomainLands(g, me, opp, "Forest", "Island")
	if got := effectstest.SelfPricedMV(t, g, me, b519Draco, tl, "{16}"); got != 10 {
		t.Errorf("three basic land types: %d, want 10", got)
	}
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Dual", TypeLine: "Land — Swamp Mountain"})
	if got := effectstest.SelfPricedMV(t, g, me, b519Draco, tl, "{16}"); got != 6 {
		t.Errorf("all five: %d, want 6", got)
	}
}

func b519Upkeep(t *testing.T, types []string, pool int, pay bool) (kept bool, prompted bool, g *game.Game) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if len(types) > 0 {
		effectstest.DomainLands(g, me, opp, types...)
	}
	draco := effectstest.B12Push(g, me.ID, "Draco", "Artifact Creature — Dragon", b519Draco, 9, 9)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FillPool(me, pool)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		prompted = true
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
	}
	return g.Battlefield.Contains(draco), prompted, g
}

func TestS82OldB519DracoUpkeepPaymentShrinksWithDomain(t *testing.T) {
	// Domain 0 would cost {10}: with only {6} in the pool it can't be paid.
	if kept, prompted, _ := b519Upkeep(t, nil, 6, true); kept {
		t.Errorf("an unpayable {10} kept Draco (prompted=%v)", prompted)
	}
	// Three types (DomainLands adds Forest, Plains, Island here via the opponent's
	// Swamp not counting): cost {4}.
	if kept, prompted, _ := b519Upkeep(t, []string{"Forest", "Island"}, 4, true); !kept || !prompted {
		t.Errorf("paying {4} with three types: kept=%v prompted=%v", kept, prompted)
	}
	if kept, _, _ := b519Upkeep(t, []string{"Forest", "Island"}, 4, false); kept {
		t.Error("declining the payment kept Draco")
	}
}

func TestS82OldB519DracoWithFiveTypesPaysNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.DomainLands(g, me, opp, "Forest", "Island")
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Dual", TypeLine: "Land — Swamp Mountain"})
	draco := effectstest.B12Push(g, me.ID, "Draco", "Artifact Creature — Dragon", b519Draco, 9, 9)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a {0} upkeep cost asked for a payment")
	}
	if !g.Battlefield.Contains(draco) {
		t.Error("Draco was sacrificed with a {0} cost")
	}
	_ = uuid.Nil
}
