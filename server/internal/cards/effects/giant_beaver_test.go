package effects

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// giant_beaver_test.go -- #3546, ADR 0186 slice 4. Giant Beaver's attack
// trigger targets a creature that saddled IT this turn: the clause is
// TargetSpec.PaidForSource, judged in the one shared walk.

const giantBeaverOracle = "2a4b42b3-6a95-41e1-9741-b35723d8da73"

// beaverTable builds the common scene: a Giant Beaver saddled by `a` and
// `b` (power 2 each, Saddle 3), a second Mount saddled by `c`, and an
// untapped bystander `idle`. The Beaver is left attacking, its trigger
// waiting on the target prompt.
type beaverScene struct {
	g           *game.Game
	me, opp     *game.Player
	beaver      uuid.UUID
	other       uuid.UUID
	a, b, c     uuid.UUID
	idle        uuid.UUID
	declaredDef uuid.UUID
}

func beaverBeforeAttack(t *testing.T) beaverScene {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	s := beaverScene{g: g, me: me, opp: opp, declaredDef: opp.ID}
	s.beaver = pushMountForTest(g, me.ID, "Giant Beaver", giantBeaverOracle, 4)
	s.other = pushMountForTest(g, me.ID, "Gilded Ghoda", gildedGhodaOracle, 2)
	s.a = pushCrewerForTest(g, me.ID, "Saddler A", 2)
	s.b = pushCrewerForTest(g, me.ID, "Saddler B", 2)
	s.c = pushCrewerForTest(g, me.ID, "Saddler C", 2)
	s.idle = pushCrewerForTest(g, me.ID, "Idle", 2)

	toMain(t, g)
	if err := saddleAbility(t, g, me.ID, s.other, s.c); err != nil {
		t.Fatalf("saddle the other Mount: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := saddleAbility(t, g, me.ID, s.beaver, s.a, s.b); err != nil {
		t.Fatalf("saddle the Beaver: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !saddleCard(t, g, s.beaver).Saddled {
		t.Fatal("setup: the Beaver is not saddled")
	}
	return s
}

func beaverScenePromptReady(t *testing.T) beaverScene {
	t.Helper()
	s := beaverBeforeAttack(t)
	declareAttack(t, s.g, s.opp.ID, s.beaver)
	return s
}

func TestGiantBeaverIsFullWithNoCaveat(t *testing.T) {
	spec, ok := Lookup(giantBeaverOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Giant Beaver: want complete with no caveat, got %+v", spec)
	}
}

// The prompt offers exactly the creatures that saddled the Beaver: not
// the creature that saddled a different Mount, not an idle bystander,
// not the Beaver or the other Mount.
func TestGiantBeaverOffersExactlyItsSaddlers(t *testing.T) {
	s := beaverScenePromptReady(t)
	pick := latestPickTarget(s.g, s.me.ID)
	if pick == nil {
		t.Fatal("no pick_target prompt for the attack trigger")
	}
	got := slices.Clone(pick.PickTargetCards)
	want := []uuid.UUID{s.a, s.b}
	if len(got) != 2 || !slices.Contains(got, s.a) || !slices.Contains(got, s.b) {
		t.Fatalf("offered %v, want exactly %v", got, want)
	}
	for name, id := range map[string]uuid.UUID{"the other Mount's saddler": s.c, "an idle creature": s.idle, "the Beaver": s.beaver, "the other Mount": s.other} {
		if slices.Contains(got, id) {
			t.Errorf("%s was offered", name)
		}
	}

	pickTriggerTarget(t, s.g, s.me.ID, s.b)
	passPriorityAroundTable(t, s.g)
	if n := countersOf(t, s.g, s.b); n != 1 {
		t.Errorf("the chosen saddler got %d counters, want 1", n)
	}
	for name, id := range map[string]uuid.UUID{"the other saddler": s.a, "the other Mount's saddler": s.c, "the Beaver": s.beaver} {
		if n := countersOf(t, s.g, id); n != 0 {
			t.Errorf("%s got %d counters, want 0", name, n)
		}
	}
}

// Announcing a creature that did not saddle the Beaver is refused (CR
// 601.2c / 603.3d), and the prompt is still open for a legal answer.
func TestGiantBeaverRefusesAnIllegalTarget(t *testing.T) {
	s := beaverScenePromptReady(t)
	pick := latestPickTarget(s.g, s.me.ID)
	if pick == nil {
		t.Fatal("no pick_target prompt")
	}
	for name, id := range map[string]uuid.UUID{"the other Mount's saddler": s.c, "an idle creature": s.idle} {
		err := s.g.ResolvePickTarget(pick.ID, s.me.ID, game.TargetRef{Kind: game.TargetCard, ID: id})
		if !errors.Is(err, game.ErrIllegalTarget) {
			t.Errorf("picking %s: err = %v, want ErrIllegalTarget", name, err)
		}
	}
	if latestPickTarget(s.g, s.me.ID) == nil {
		t.Fatal("a refused pick closed the prompt")
	}
	pickTriggerTarget(t, s.g, s.me.ID, s.a)
	passPriorityAroundTable(t, s.g)
	if n := countersOf(t, s.g, s.a); n != 1 {
		t.Errorf("the legal pick got %d counters, want 1", n)
	}
}

// A saddler flickered in response is a new object that saddled nothing:
// the trigger's target is illegal at resolution (CR 608.2b) and nothing
// gets a counter.
func TestGiantBeaverFizzlesWhenItsTargetIsFlickeredInResponse(t *testing.T) {
	s := beaverScenePromptReady(t)
	pickTriggerTarget(t, s.g, s.me.ID, s.a)
	if triggerOnStack(s.g, s.beaver) == nil {
		t.Fatal("the attack trigger is not on the stack")
	}
	flickerForTest(t, s.g, s.me.ID, s.a)
	passPriorityAroundTable(t, s.g)

	for _, c := range s.g.Battlefield.Cards {
		if c.Counters[game.CounterPlusOne] != 0 {
			t.Errorf("%s got a +1/+1 counter from a fizzled trigger", c.Name)
		}
	}
	if triggerOnStack(s.g, s.beaver) != nil {
		t.Error("the trigger is still on the stack")
	}
}

// The Beaver bouncing in response does not change who saddled the
// object the trigger came from: its target is read from the departed
// source's record, and still resolves.
func TestGiantBeaverTriggerResolvesAfterTheBeaverLeaves(t *testing.T) {
	s := beaverScenePromptReady(t)
	pickTriggerTarget(t, s.g, s.me.ID, s.a)
	s.g.WithWriteLock(func() {
		if err := s.g.BounceToHandForEffect(s.beaver); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	passPriorityAroundTable(t, s.g)
	if n := countersOf(t, s.g, s.a); n != 1 {
		t.Errorf("the saddler got %d counters after the Beaver left, want 1", n)
	}
}

// With no saddler left on the battlefield the trigger is removed for want
// of a target (CR 603.3d): no prompt, nothing on the stack.
func TestGiantBeaverTriggerIsRemovedWithNoSaddlerOnTheBattlefield(t *testing.T) {
	s := beaverBeforeAttack(t)
	for _, id := range []uuid.UUID{s.a, s.b} {
		s.g.WithWriteLock(func() {
			if err := s.g.BounceToHandForEffect(id); err != nil {
				t.Fatalf("bounce: %v", err)
			}
		})
	}
	declareAttack(t, s.g, s.opp.ID, s.beaver)

	if latestPickTarget(s.g, s.me.ID) != nil {
		t.Error("a target prompt was queued with no legal target")
	}
	if triggerOnStack(s.g, s.beaver) != nil {
		t.Error("the trigger is on the stack with no legal target")
	}
	// The creature that saddled the OTHER Mount is not a stand-in.
	if n := countersOf(t, s.g, s.c); n != 0 {
		t.Errorf("the other Mount's saddler got %d counters", n)
	}
}

// A creature that saddled the Beaver and then left and came back under
// the same card (new object) is not offered (CR 400.7).
func TestGiantBeaverDoesNotOfferAReturnedSaddler(t *testing.T) {
	s := beaverBeforeAttack(t)
	flickerForTest(t, s.g, s.me.ID, s.a)
	declareAttack(t, s.g, s.opp.ID, s.beaver)
	pick := latestPickTarget(s.g, s.me.ID)
	if pick == nil {
		t.Fatal("no prompt: B should still be a legal target")
	}
	if len(pick.PickTargetCards) != 1 || pick.PickTargetCards[0] != s.b {
		t.Errorf("offered %v, want only %s", pick.PickTargetCards, s.b)
	}
}
