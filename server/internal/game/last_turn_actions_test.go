package game

import (
	"testing"

	"github.com/google/uuid"
)

// last_turn_actions_test.go — #1997 (Arboria): who PUT a permanent onto
// the battlefield, and whether a player cast a spell or put a nontoken
// permanent onto the battlefield during their last turn.

func putCount(g *Game, p *Player) (n int) {
	g.ReadSnapshot(func() { n = g.TurnTallyFor(p.ID).NontokenPermanentsPut })
	return n
}

func castOrPutLastTurn(g *Game, p *Player) (ok bool) {
	g.ReadSnapshot(func() { ok = g.CastOrPutDuringLastTurnForEffect(p.ID) })
	return ok
}

// handPermanentForTest puts a vanilla permanent card in p's hand.
func handPermanentForTest(p *Player, name, typeLine string) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine = typeLine
	c.Controller = p.ID
	p.Hand.PushTop(c)
	return c.InstanceID
}

// resolveTriggerForTest puts a trigger controlled by `controller` on the
// stack with `effect` as its resolution, and resolves the stack.
func resolveTriggerForTest(t *testing.T, g *Game, controller *Player, effect func(g *Game, item *StackItem) error) {
	t.Helper()
	item := &StackItem{
		ID:           uuid.New(),
		Kind:         StackItemTriggered,
		Controller:   controller.ID,
		Owner:        controller.ID,
		SourceCardID: uuid.New(),
		Label:        "Test trigger — put a permanent onto the battlefield",
		Effect:       effect,
	}
	g.WithWriteLock(func() {
		g.PendingTriggers = append(g.PendingTriggers, item)
		g.drainPendingTriggersAPNAPLocked()
	})
	resolveWholeStackForTest(t, g)
}

// An effect's put is made by the controller of the resolving ability
// (CR 608.2c), not by the player the permanent enters under; inside
// PutByEnteringPlayerForEffect ("each player puts", CR 110.2a) it is the
// player it enters under.
func TestPutterIsTheResolvingControllerUnlessTheInstructionNamesAnother(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	first := handPermanentForTest(b, "Bear", "Creature — Bear")
	resolveTriggerForTest(t, g, a, func(g *Game, _ *StackItem) error {
		return g.PutCardsFromHandOntoBattlefieldKeyedForEffect([]uuid.UUID{first}, HandEntryOptions{}, KeyedTail{})
	})
	if c := findBattlefieldCard(g, first); c == nil || c.Controller != b.ID {
		t.Fatal("setup: the Bear did not enter under its owner")
	}
	if got := putCount(g, a); got != 1 {
		t.Errorf("the ability's controller put %d permanents, want 1 (CR 608.2c)", got)
	}
	if got := putCount(g, b); got != 0 {
		t.Errorf("the player it entered under put %d, want 0: they did not put it", got)
	}

	second := handPermanentForTest(b, "Elk", "Creature — Elk")
	resolveTriggerForTest(t, g, a, func(g *Game, _ *StackItem) error {
		return g.PutByEnteringPlayerForEffect(func() error {
			return g.PutCardsFromHandOntoBattlefieldKeyedForEffect([]uuid.UUID{second}, HandEntryOptions{}, KeyedTail{})
		})
	})
	if got := putCount(g, b); got != 1 {
		t.Errorf("an instruction naming the player: they put %d, want 1 (CR 110.2a)", got)
	}
	if got := putCount(g, a); got != 1 {
		t.Errorf("the ability's controller is still at %d, want 1", got)
	}
}

// A land play is its player's, a permanent spell its caster's; a token
// is not a nontoken permanent; a manifest is the manifesting player's, and
// a search the searcher's.
func TestPutterOfALandPlayASpellATokenAManifestAndASearch(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)
	land := handPermanentForTest(a, "Wastes", "Basic Land")
	if err := g.CastSpell(a.ID, land, CastSpellParams{}); err != nil {
		t.Fatalf("land play: %v", err)
	}
	if got := putCount(g, a); got != 1 {
		t.Fatalf("a land play put %d, want 1 (CR 305.1)", got)
	}
	spell := NewCard("Free Relic", a.ID)
	spell.TypeLine, spell.ManaCost, spell.Controller = "Artifact", "{0}", a.ID
	a.Hand.PushTop(spell)
	if err := g.CastSpell(a.ID, spell.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	resolveWholeStackForTest(t, g)
	if got := putCount(g, a); got != 2 {
		t.Fatalf("a resolved permanent spell: %d, want 2", got)
	}
	resolveTriggerForTest(t, g, a, func(g *Game, _ *StackItem) error {
		return g.CreateTokenForEffect(a.ID, Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1}, 2)
	})
	if got := putCount(g, a); got != 2 {
		t.Errorf("tokens counted: %d, want still 2", got)
	}
	// "Its controller manifests": the trigger is a's, the manifest b's.
	resolveTriggerForTest(t, g, a, func(g *Game, _ *StackItem) error {
		_, err := g.ManifestForEffect(b.ID)
		return err
	})
	if got := putCount(g, b); got != 1 {
		t.Errorf("the manifesting player put %d, want 1 (CR 701.40a)", got)
	}
	if got := putCount(g, a); got != 2 {
		t.Errorf("the trigger's controller is at %d after another player's manifest, want 2", got)
	}
	// "Its controller may search their library … put that card onto the
	// battlefield" (Path to Exile): the searcher puts it.
	basic := NewCard("Wastes", b.ID)
	basic.TypeLine, basic.Controller = "Basic Land", b.ID
	b.Library.PushTop(basic)
	resolveTriggerForTest(t, g, a, func(g *Game, _ *StackItem) error {
		return g.SearchLibraryForEffect(b.ID, func(c Card) bool { return c.Name == "Wastes" }, ZoneBattlefield, 1, false, true)
	})
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == b.ID {
			if err := g.ResolveSearchLibrary(c.ID, b.ID, []uuid.UUID{basic.InstanceID}); err != nil {
				t.Fatalf("ResolveSearchLibrary: %v", err)
			}
			break
		}
	}
	if !g.Battlefield.Contains(basic.InstanceID) {
		t.Fatal("setup: the searched land did not enter")
	}
	if got := putCount(g, b); got != 2 {
		t.Errorf("the searcher put %d, want 2 (the manifest and the land)", got)
	}
	if got := putCount(g, a); got != 2 {
		t.Errorf("the trigger's controller is at %d after another player's search, want 2", got)
	}
}

// "During their last turn": written as each player's turn ends, kept
// through the other players' turns, replaced by their next one, and
// carried by a restore.
func TestCastOrPutDuringLastTurnIsTheirMostRecentTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	if castOrPutLastTurn(g, a) || castOrPutLastTurn(g, b) {
		t.Fatal("before anyone's first turn ends, nobody did anything last turn")
	}
	toMainPhase(t, g)
	land := handPermanentForTest(a, "Wastes", "Basic Land")
	if err := g.CastSpell(a.ID, land, CastSpellParams{}); err != nil {
		t.Fatalf("land play: %v", err)
	}
	if castOrPutLastTurn(g, a) {
		t.Fatal("this turn is not yet seat 0's LAST turn")
	}
	passTurn(t, g)
	if !castOrPutLastTurn(g, a) {
		t.Fatal("seat 0 played a land during its last turn")
	}
	// Seat 1 casts an instant on its own turn, and puts nothing.
	toMainPhase(t, g)
	castForTest(t, g, b, "{0}", CastSpellParams{})
	resolveWholeStackForTest(t, g)
	passTurn(t, g)
	if !castOrPutLastTurn(g, b) {
		t.Error("seat 1 cast a spell during its last turn")
	}
	passTurn(t, g)
	passTurn(t, g)
	if !castOrPutLastTurn(g, a) {
		t.Error("three other turns do not touch seat 0's last turn")
	}
	_, restored := roundTrip(t, g)
	if !castOrPutLastTurn(restored, a) || !castOrPutLastTurn(restored, b) {
		t.Error("a restore lost the record of the last turn")
	}
	passTurn(t, restored) // seat 0's turn, with nothing done
	if castOrPutLastTurn(restored, restored.Seats[0]) {
		t.Error("seat 0 did nothing during its newest turn, which replaces the old one")
	}
}
