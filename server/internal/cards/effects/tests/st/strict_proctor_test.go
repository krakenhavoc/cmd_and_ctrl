package testsst

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #6158 — Strict Proctor: "Whenever a permanent entering causes a
// triggered ability to trigger, counter that ability unless its
// controller pays {2}."

const (
	strictProctorOracle   = "b967870d-9773-4ad5-bf06-c3be5465dab9"
	strictProctorLabel    = "Strict Proctor — counter that ability"
	strictProctorTraveler = "a30907c0-fbde-4fd3-a8c7-f304305fcea7" // Doomed Traveler
)

// strictProctorEnter puts a permanent onto the battlefield and announces
// its entry, then runs the boundary that puts what it triggered on the
// stack.
func strictProctorEnter(g *game.Game, controller uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: 2, Toughness: 2, Owner: controller, Controller: controller,
	})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: id, Actor: controller})
	})
	g.RunStateChecksForTest()
	return id
}

func strictProctorItems(g *game.Game, prefix string) []*game.StackItem {
	var out []*game.StackItem
	g.WithWriteLock(func() {
		for _, it := range g.StackMeta {
			if it != nil && strings.HasPrefix(it.Label, prefix) {
				out = append(out, it)
			}
		}
		for _, it := range g.PendingTriggers {
			if it != nil && strings.HasPrefix(it.Label, prefix) {
				out = append(out, it)
			}
		}
	})
	return out
}

// resolveProctor passes priority until the Proctor's pay-unless prompt
// is addressed to payer.
func resolveProctor(t *testing.T, g *game.Game, payer uuid.UUID) {
	t.Helper()
	for i := 0; i < 16 && !effectstest.HasPayUnlessFor(g, payer); i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !effectstest.HasPayUnlessFor(g, payer) {
		t.Fatal("the Proctor's ability raised no pay-unless prompt for the ability's controller")
	}
}

// CR 603.3b: the Proctor's ability, whose trigger condition is another
// ability triggering, goes on the stack above the ETB ability even on its
// own controller's turn, where APNAP alone would put it below. Declined,
// the ETB ability is countered (CR 701.6a) and draws nothing.
func TestStrictProctorCountersAnEntryTriggerUnlessPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Strict Proctor", "Creature — Spirit Cleric", strictProctorOracle, false)
	strictProctorEnter(g, opp.ID, "Mulldrifter", "Creature — Elemental", effectstest.MulldrifterOracle)

	drift := strictProctorItems(g, "Mulldrifter")
	proctor := strictProctorItems(g, strictProctorLabel)
	if len(drift) != 1 || len(proctor) != 1 {
		t.Fatalf("want one Mulldrifter trigger and one Proctor trigger, got %d and %d", len(drift), len(proctor))
	}
	if proctor[0].Seq <= drift[0].Seq {
		t.Fatalf("the Proctor's trigger (seq %d) is not above the Mulldrifter trigger (seq %d)", proctor[0].Seq, drift[0].Seq)
	}
	if proctor[0].Controller != me.ID {
		t.Fatalf("the Proctor's trigger is controlled by %v, want %v", proctor[0].Controller, me.ID)
	}

	hand := opp.Hand.Size()
	resolveProctor(t, g, opp.ID)
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - hand; got != 0 {
		t.Errorf("the countered Mulldrifter trigger still drew %d cards", got)
	}
	if !effectstest.StackFullyEmpty(g) {
		t.Error("the stack should be empty once the countered ability is gone")
	}
}

// Paid ({2} from the ability's controller's pool), the ability resolves.
// The Proctor's own controller is asked too: "a triggered ability" is
// anyone's.
func TestStrictProctorLetsAPaidEntryTriggerResolve(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Strict Proctor", "Creature — Spirit Cleric", strictProctorOracle, false)
	strictProctorEnter(g, me.ID, "Mulldrifter", "Creature — Elemental", effectstest.MulldrifterOracle)
	if n := len(strictProctorItems(g, strictProctorLabel)); n != 1 {
		t.Fatalf("your own permanent's entry trigger should trigger the Proctor once, got %d", n)
	}
	hand := me.Hand.Size()
	resolveProctor(t, g, me.ID)
	effectstest.FloatForTest(g, me, "CC")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("the paid-for Mulldrifter trigger drew %d cards, want 2", got)
	}
}

// A dies trigger is not caused by a permanent entering: the Proctor
// does not trigger and Doomed Traveler's Spirit is made.
func TestStrictProctorIgnoresADiesTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Strict Proctor", "Creature — Spirit Cleric", strictProctorOracle, false)
	traveler := effectstest.PushCatalogPermanent(g, opp.ID, "Doomed Traveler", "Creature — Human Soldier", strictProctorTraveler, false)
	effectstest.Destroy(t, g, traveler)
	g.RunStateChecksForTest()
	if n := len(strictProctorItems(g, strictProctorLabel)); n != 0 {
		t.Fatalf("a dies trigger triggered the Proctor (%d)", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("a dies trigger was charged {2}")
	}
	if n := effectstest.CountOnBattlefield(g, "Spirit", opp.ID); n != 1 {
		t.Errorf("Doomed Traveler made %d Spirits, want 1", n)
	}
}
