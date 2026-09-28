package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// goad_multi_test.go — #1598 through the card path: a creature Alela
// goads keeps an earlier or later goad by somebody else (CR 701.15c),
// so Alela, Cunning Conqueror ships without her one-goader caveat; and
// the "the goad ends" delayed trigger every goad card still schedules
// ends nothing the engine has not already ended.

// TestAlelasGoadKeepsAnotherPlayersGoad — Alela's controller goads the
// victim's Bear, then another opponent goads it too. On the victim's
// turn the Bear may attack neither goader and must attack the fourth
// seat. Each goad then ends on its own goader's turn.
func TestAlelasGoadKeepsAnotherPlayersGoad(t *testing.T) {
	g := newCatalogGame(t)
	me, victim, other, fourth := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	alela := b12Push(g, me.ID, "Alela, Cunning Conqueror", "Legendary Creature — Faerie Warlock", b33AlelaOracle, 2, 4)
	theirs := b12Creature(g, victim.ID, "Their Bear", "Creature — Bear", 2, 2)
	attackWith(t, g, victim.ID, alela)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { b33Goad(g, theirs, other.ID) })

	c, _ := battlefieldCard(g, theirs)
	if got := c.Goaders(); len(got) != 2 || got[0] != me.ID || got[1] != other.ID {
		t.Fatalf("goaders = %v, want Alela's controller then the other opponent", got)
	}

	advanceToDeclareAttackersOf(t, g, 1)
	for _, goader := range []*game.Player{me, other} {
		if err := g.DeclareAttacker(theirs, goader.ID); !errors.Is(err, game.ErrAttackRequirement) {
			t.Fatalf("the twice-goaded Bear at a goader = %v, want ErrAttackRequirement", err)
		}
	}
	if err := g.DeclareAttacker(theirs, fourth.ID); err != nil {
		t.Fatalf("the twice-goaded Bear at the player who goaded it neither time: %v", err)
	}

	advanceToUpkeepOf(t, g, 2)
	if c, _ := battlefieldCard(g, theirs); !c.IsGoadedBy(me.ID) || c.IsGoadedBy(other.ID) {
		t.Fatalf("on the other opponent's turn: goaders = %v, want Alela's controller alone", c.Goaders())
	}
}

// TestGoadEndsTriggerSparesARefreshedGoad — the legacy "the goad ends"
// trigger lands at the goader's upkeep, after the engine has already
// ended the goad as that turn began. A goad the same player makes in
// that upkeep, before the trigger resolves, is for their NEXT turn,
// and the trigger must not take it.
func TestGoadEndsTriggerSparesARefreshedGoad(t *testing.T) {
	g := newCatalogGame(t)
	me, victim, other := g.Seats[0], g.Seats[1], g.Seats[2]
	alela := b12Push(g, me.ID, "Alela, Cunning Conqueror", "Legendary Creature — Faerie Warlock", b33AlelaOracle, 2, 4)
	theirs := b12Creature(g, victim.ID, "Their Bear", "Creature — Bear", 2, 2)
	attackWith(t, g, victim.ID, alela)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)

	advanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(theirs, other.ID); err != nil {
		t.Fatalf("the goaded Bear at another opponent: %v", err)
	}
	advanceToUpkeepOf(t, g, 0)
	if c, _ := battlefieldCard(g, theirs); c.IsGoaded() {
		t.Fatalf("the goad outlived the start of its goader's turn: %v", c.Goaders())
	}
	if acItemLabelled(g, "Alela, Cunning Conqueror — the goad ends") == uuid.Nil {
		t.Fatal("the legacy goad-ends trigger is not on the stack at the goader's upkeep")
	}
	g.WithWriteLock(func() { b33Goad(g, theirs, me.ID) })
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, theirs); !c.IsGoadedBy(me.ID) {
		t.Fatal("the goad-ends trigger took a goad made after the goader's turn began")
	}
}
