package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 GAP2 batch 04 (worker a2): Rampaging War Mammoth, the trigger half
// of #6085 — "up to X target artifacts" sized by the cycling cost's X.

const s82g2a2Mammoth = "05e910aa-cf84-468c-8a48-256588cbbf08"

func TestS82Gap2B04A2MammothIsFullWithNoCaveats(t *testing.T) {
	spec, ok := effects.Lookup(s82g2a2Mammoth)
	if !ok {
		t.Fatal("Rampaging War Mammoth is not registered")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("%v with caveats %v, want full with none", spec.Completeness, spec.Caveats)
	}
}

func s82g2a2CycleMammoth(t *testing.T, g *game.Game, me *game.Player, x int) {
	t.Helper()
	id := effectstest.PushCatalogHandCard(me, "Rampaging War Mammoth", "Creature — Elephant", s82g2a2Mammoth)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Drawn"})
	effectstest.FillPool(me, x+2)
	effectstest.FillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, XValue: x}); err != nil {
		t.Fatalf("cycle at X=%d: %v", x, err)
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("the Mammoth was not discarded")
	}
}

// X=2: the prompt offers the artifacts (not the creature) and takes up to
// two; three picks are refused; the two chosen are destroyed.
func TestS82Gap2B04A2MammothDestroysUpToXArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := effectstest.PushCatalogPermanent(g, opp.ID, "Rock A", "Artifact", "", false)
	b := effectstest.PushCatalogPermanent(g, opp.ID, "Rock B", "Artifact", "", false)
	c := effectstest.PushCatalogPermanent(g, opp.ID, "Rock C", "Artifact", "", false)
	bear := effectstest.PushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	s82g2a2CycleMammoth(t, g, me, 2)

	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt for the cycling trigger")
	}
	if prompt.PickTargetMin != 0 || prompt.PickTargetMax != 2 {
		t.Errorf("prompt takes %d..%d targets, want 0..2", prompt.PickTargetMin, prompt.PickTargetMax)
	}
	if effectstest.HasID(prompt.PickTargetCards, bear) || !effectstest.HasID(prompt.PickTargetCards, c) {
		t.Errorf("offered %v: want the artifacts only", prompt.PickTargetCards)
	}
	three := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c}}
	if err := g.ResolvePickTargets(prompt.ID, me.ID, three); err == nil {
		t.Fatal("three targets accepted for X=2")
	}
	if err := g.ResolvePickTargets(prompt.ID, me.ID, three[:2]); err != nil {
		t.Fatalf("two targets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if _, ok := effectstest.BattlefieldCard(g, id); ok {
			t.Error("a chosen artifact survived")
		}
	}
	if _, ok := effectstest.BattlefieldCard(g, c); !ok {
		t.Error("the unchosen artifact was destroyed")
	}
}

// X=0: no target to choose at all (never an unbounded clause), and the
// cycling still draws.
func TestS82Gap2B04A2MammothAtXZeroTargetsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := effectstest.PushCatalogPermanent(g, opp.ID, "Rock A", "Artifact", "", false)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	hand := me.Hand.Size()
	s82g2a2CycleMammoth(t, g, me, 0)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatalf("X=0 opened a target prompt for %d..%d targets", p.PickTargetMin, p.PickTargetMax)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, a); !ok {
		t.Error("X=0 destroyed an artifact")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (the Mammoth pushed, discarded, one drawn)", me.Hand.Size(), hand+1)
	}
}
