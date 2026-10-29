package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// recruit_test.go — the recruit keyword action (CR 701.70a, #4475):
// draw a card, then discard a card; a nonland card discarded makes a 1/1
// white Human Soldier, a land discarded makes nothing, and the discard is
// a restore point.

// startRecruit empties the seat's hand, puts `top` on its library and
// recruits for it from a bare triggered context.
func startRecruit(t *testing.T, g *game.Game, me *game.Player, topName, topType string) uuid.UUID {
	t.Helper()
	var drawn uuid.UUID
	var err error
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		drawn = seedLibraryTop(me, topName, topType, "")
		err = Recruit{}.Apply(NewContext(g, &game.StackItem{
			Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID,
		}))
	})
	if err != nil {
		t.Fatalf("Recruit: %v", err)
	}
	return drawn
}

func recruitSoldiers(g *game.Game, me uuid.UUID) int {
	n, _ := countTokensNamed(g, me, recruitSoldier().Name)
	return n
}

// A nonland card discarded makes the Human Soldier, under the recruiter.
func TestRecruitDiscardingANonlandCardMakesAHumanSoldier(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	drawn := startRecruit(t, g, me, "Drawn Spell", "Sorcery")
	if got := me.Hand.Size(); got != 1 {
		t.Fatalf("hand after the draw has %d cards, want 1", got)
	}
	sweep01Discard(t, g, me.ID, recruitDiscarded.Key())
	answerDiscard(t, g, me.ID, drawn)
	if got := recruitSoldiers(g, me.ID); got != 1 {
		t.Fatalf("%d Human Soldier tokens after discarding a sorcery, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if IsToken(c) && c.Controller == me.ID {
			if !c.IsCreature() || c.Power != 1 || c.Toughness != 1 || !c.HasSubtype("Human") || !c.HasSubtype("Soldier") ||
				len(c.Colors) != 1 || c.Colors[0] != "W" {
				t.Errorf("the token is %q %d/%d %v, want a 1/1 white Human Soldier creature", c.TypeLine, c.Power, c.Toughness, c.Colors)
			}
		}
	}
}

// A land discarded makes nothing.
func TestRecruitDiscardingALandMakesNoToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	drawn := startRecruit(t, g, me, "Drawn Land", "Land")
	answerDiscard(t, g, me.ID, drawn)
	if got := recruitSoldiers(g, me.ID); got != 0 {
		t.Errorf("%d Human Soldier tokens after discarding a land, want 0", got)
	}
}

// The player chooses: with a land and a spell in hand, discarding the
// land when the spell was available makes nothing, and the spell stays.
func TestRecruitTheDiscardIsTheRecruitersChoice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	var spell uuid.UUID
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		spell = handCardFull(me, "Kept Spell", "Instant", "{U}", "", nil)
	})
	drawn := seedLibraryTop(me, "Drawn Land", "Land", "")
	g.WithWriteLock(func() {
		_ = Recruit{}.Apply(NewContext(g, &game.StackItem{Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID}))
	})
	answerDiscard(t, g, me.ID, drawn)
	if got := recruitSoldiers(g, me.ID); got != 0 {
		t.Errorf("%d tokens after choosing the land, want 0", got)
	}
	if !me.Hand.Contains(spell) {
		t.Error("the spell left the hand; only the chosen card is discarded")
	}
}

// The discard is a restore point: the restored table's answer still
// makes the token.
func TestRecruitDiscardIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	drawn := startRecruit(t, g, me, "Drawn Spell", "Creature — Elf")
	sweep01Discard(t, g, me.ID, recruitDiscarded.Key())

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	answerDiscard(t, g, me.ID, drawn)
	if got := recruitSoldiers(g, me.ID); got != 1 {
		t.Errorf("%d Human Soldier tokens on the restored table, want 1", got)
	}
}

// Nothing to draw and nothing in hand: no discard, no token (CR 701.9a).
func TestRecruitWithNothingToDiscardMakesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	var err error
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		me.Library.Cards = nil
		err = Recruit{}.Apply(NewContext(g, &game.StackItem{Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID}))
	})
	if err != nil {
		t.Fatalf("Recruit: %v", err)
	}
	if c := discardChoiceFor(g, me.ID); c != nil {
		t.Errorf("a discard prompt opened on an empty hand: %+v", c)
	}
	if got := recruitSoldiers(g, me.ID); got != 0 {
		t.Errorf("%d tokens from an empty recruit, want 0", got)
	}
}
