package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82g24Hullbreacher = "2f533667-e29b-4bec-897a-e7a9eee08314"
	s82g24SearchElem   = "f3d41f6d-1792-4a83-a4e8-d1b6d3ea04a8"
	s82g24Subgoyf      = "9e57331c-b5b0-40b7-ba2d-6effa9d0abea"
)

func TestS82GapB24EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82g24Hullbreacher: "Hullbreacher", s82g24SearchElem: "Search Elemental", s82g24Subgoyf: "Subgoyf",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82GapB24HullbreacherTurnsOpponentDrawsIntoTreasures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedReplacementPermanent(g, s82g24Hullbreacher, "Hullbreacher", me.ID)
	mine, theirs := me.Hand.Size(), opp.Hand.Size()

	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 2) })
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != theirs || me.Hand.Size() != mine {
		t.Errorf("hands changed: opp %d → %d, me %d → %d; the opponent's draws are replaced and drew nothing",
			theirs, opp.Hand.Size(), mine, me.Hand.Size())
	}
	if got := battlefieldCountNamed(g, me.ID, "Treasure"); got != 2 {
		t.Errorf("%d Treasure tokens for two replaced draws, want 2", got)
	}
	if got := battlefieldCountNamed(g, opp.ID, "Treasure"); got != 0 {
		t.Errorf("the opponent got %d Treasures", got)
	}

	// My own draw is untouched.
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	if me.Hand.Size() != mine+1 {
		t.Errorf("my own draw was replaced: hand %d, want %d", me.Hand.Size(), mine+1)
	}

	// The opponent's first draw of their own draw step is left alone.
	mine, theirs = me.Hand.Size(), opp.Hand.Size()
	advanceToUpkeepOf(t, g, 1)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	if g.Turn.Step != game.StepDraw {
		t.Fatalf("expected the draw step, at %s", g.Turn.Step)
	}
	if opp.Hand.Size() != theirs+1 || me.Hand.Size() != mine {
		t.Errorf("draw step: opp %d → %d (want +1), me %d → %d (want same)", theirs, opp.Hand.Size(), mine, me.Hand.Size())
	}
	if got := battlefieldCountNamed(g, me.ID, "Treasure"); got != 2 {
		t.Errorf("%d Treasures after the draw step, want still 2", got)
	}
}

func TestS82GapB24SearchElementalScriesOnYourSearchAndGrowsOnAnyScry(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	elemental := b12Push(g, me.ID, "Search Elemental", "Creature — Elemental", s82g24SearchElem, 1, 1)

	// An opponent's search does nothing.
	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(opp.ID, func(game.Card) bool { return true }, game.ZoneHand, 1, false, false)
	})
	answerSearchByID(t, g, opp.ID, opp.Library.Cards[0].InstanceID)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) != nil || countersOn(g, elemental, game.CounterPlusOne) != 0 {
		t.Fatal("an opponent's search triggered the Elemental")
	}

	// My search scries 1, and that scry grows it.
	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(me.ID, func(game.Card) bool { return true }, game.ZoneHand, 1, false, false)
	})
	answerSearchByID(t, g, me.ID, me.Library.Cards[0].InstanceID)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatalf("no scry prompt after my search: %+v", g.PendingChoices)
	}
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, elemental, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters after the scry, want 1", got)
	}
	if p, tough := ptOf(t, g, elemental); p != 2 || tough != 2 {
		t.Errorf("Elemental is %d/%d, want 2/2", p, tough)
	}
	if !cardHasRestriction(t, g, elemental, game.CantBeBlocked) {
		t.Error("the Elemental can still be blocked")
	}
}

// cardHasRestriction reports whether the permanent's effective state carries the restriction.
func cardHasRestriction(t *testing.T, g *game.Game, id uuid.UUID, r game.Restriction) bool {
	t.Helper()
	return layeredCard(t, g, id).Effective().Restrictions&r != 0
}

func TestS82GapB24SubgoyfCountsDifferentNonCreatureSubtypesInAllGraveyards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	goyf := b12Push(g, me.ID, "Subgoyf", "Artifact Creature — Lhurgoyf", s82g24Subgoyf, 0, 0)
	if p, tough := ptOf(t, g, goyf); p != 0 || tough != 1 {
		t.Fatalf("with empty graveyards Subgoyf is %d/%d, want 0/1", p, tough)
	}
	put := func(p *game.Player, name, typeLine string) {
		id := uuid.New()
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: p.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: p.ID}, id); err != nil {
			t.Fatal(err)
		}
	}
	// A creature type adds nothing; neither does a card with no subtype.
	put(me, "Bear", "Creature — Bear")
	put(opp, "Plain Sorcery", "Sorcery")
	if p, tough := ptOf(t, g, goyf); p != 0 || tough != 1 {
		t.Errorf("creature types and bare card types counted: %d/%d, want 0/1", p, tough)
	}
	// Plains, Equipment and Aura are three different subtypes, across two graveyards.
	put(me, "Plains", "Basic Land — Plains")
	put(opp, "Sword", "Artifact — Equipment")
	put(opp, "Pacifism", "Enchantment — Aura")
	if p, tough := ptOf(t, g, goyf); p != 3 || tough != 4 {
		t.Errorf("Subgoyf is %d/%d, want 3/4", p, tough)
	}
	// The same subtype on another card counts once; a creature type on a nonland card still doesn't.
	put(me, "Another Plains", "Basic Land — Plains")
	put(me, "Tribal Elf Spell", "Tribal Instant — Elf")
	if p, tough := ptOf(t, g, goyf); p != 3 || tough != 4 {
		t.Errorf("after a repeated subtype Subgoyf is %d/%d, want still 3/4", p, tough)
	}
}
