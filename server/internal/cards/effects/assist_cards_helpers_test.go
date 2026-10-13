package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// assist_cards_helpers_test.go — shared helpers for the assist cards
// catalogued under #3204 (one card file, one test file, one oracle
// fixture each). Each test casts its card for real and drives the stack
// with the same pass-priority loop the rest of the package uses; the
// helpers here only answer the prompts a card's own text raises.

// assistCardSpec checks the card is catalogued as Full and prints
// assist, read from the catalog the way the engine reads it.
func assistCardSpec(t *testing.T, oracle, name string) {
	t.Helper()
	spec, ok := Lookup(oracle)
	if !ok {
		t.Fatalf("%s is not registered", name)
	}
	if spec.Name != name || spec.Completeness != CompletenessFull {
		t.Fatalf("%s registered as %q with completeness %v, want Full", name, spec.Name, spec.Completeness)
	}
	c := game.NewCard(name, uuid.Nil)
	c.OracleID = oracle
	if !game.HasAssist(&c) {
		t.Errorf("%s does not read as having assist", name)
	}
}

// assistCardCastXErr casts an {X} spell from the active seat's hand
// with X announced as x, paying from the pool colours given, at the
// given targets. It returns the cast's error, so a refusal can be
// asserted.
func assistCardCastXErr(t *testing.T, g *game.Game, name, typeLine, oracle, manaCost string, pool []string, x int, targets []game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Owner: active.ID, Controller: active.ID,
	})
	advanceToMain(t, g)
	apaMana(active, pool...)
	return id, g.CastSpell(active.ID, id, game.CastSpellParams{Targets: targets, XValue: x, Strict: true})
}

// assistCardCastX is assistCardCastXErr for a cast that must succeed.
func assistCardCastX(t *testing.T, g *game.Game, name, typeLine, oracle, manaCost string, pool []string, x int, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	id, err := assistCardCastXErr(t, g, name, typeLine, oracle, manaCost, pool, x, targets)
	if err != nil {
		t.Fatalf("CastSpell %s (X=%d): %v", name, x, err)
	}
	return id
}

// assistCardSettle passes priority until the stack is empty. A
// pick_target prompt for chooser is answered with pick when one is
// given; an optional trigger prompt for chooser is answered with yes.
func assistCardSettle(t *testing.T, g *game.Game, chooser uuid.UUID, pick *game.TargetRef, yes bool) {
	t.Helper()
	for i := 0; i < 16; i++ {
		// A trigger waiting on its target is off the stack but still
		// open, so the open prompts are checked before the stack.
		if p := latestPickTarget(g, chooser); p != nil {
			if pick == nil {
				t.Fatalf("an unexpected pick_target prompt is open for %s", chooser)
			}
			if err := g.ResolvePickTarget(p.ID, p.Chooser, *pick); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if c := assistCardTriggerPrompt(g, chooser); c != nil {
			if err := g.ResolveTriggerPrompt(c.ID, c.Chooser, yes); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
			continue
		}
		if stackFullyEmpty(g) {
			return
		}
		passPriorityAroundTable(t, g)
	}
	t.Fatal("the stack did not settle")
}

// assistCardTriggerPrompt returns an open optional-trigger prompt for
// chooser, or nil.
func assistCardTriggerPrompt(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// assistCardInZone reports whether card id is in the zone of kind k,
// wherever the engine currently has it.
func assistCardInZone(g *game.Game, id uuid.UUID, k game.ZoneKind) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == k
}

// assistCardOnBattlefield reports whether card id is on the battlefield.
func assistCardOnBattlefield(g *game.Game, id uuid.UUID) bool {
	return assistCardInZone(g, id, game.ZoneBattlefield)
}

// assistCardTapped sets a battlefield creature's tapped flag for the
// test's setup, through the same card the engine reads.
func assistCardTapped(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	c := findBattlefieldCardForTest(g, id)
	if c == nil {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	c.Tapped = true
}
