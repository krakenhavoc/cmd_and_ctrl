package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// phylactery_lich_test.go — #1893, ADR 0170: an as-enters choice of
// another permanent, and the state trigger that watches it.

const phylacteryLichOracle = "520c4d88-34b6-4a26-81ca-65eb8b10461c"

func castPhylacteryLich(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, "Phylactery Lich", "Creature — Zombie", phylacteryLichOracle, nil)
	passPriorityAroundTable(t, g)
	return id
}

func TestPhylacteryLichPutsTheCounterBeforeItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	golem := b12Creature(g, me.ID, "Steel Golem", "Artifact Creature — Golem", 3, 4)
	rock := b12Permanent(g, me.ID, "Mind Stone", "Artifact")
	theirs := b12Permanent(g, opp.ID, "Their Stone", "Artifact")
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")

	id := castPhylacteryLich(t, g)
	c := entryChoiceOfKind(g, game.PendingChoiceEntryChoosePermanents, me.ID)
	if c == nil {
		t.Fatal("the Lich did not ask for an artifact")
	}
	if _, ok := battlefieldCard(g, id); ok {
		t.Fatal("the Lich entered before the choice was made (CR 614.12)")
	}
	if c.ChooseMin != 1 || c.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 1..1 (not a may)", c.ChooseMin, c.ChooseMax)
	}
	offered := map[uuid.UUID]bool{}
	for _, cand := range c.ChooseCards {
		offered[cand] = true
	}
	if !offered[golem] || !offered[rock] || offered[theirs] || len(offered) != 2 {
		t.Errorf("candidates %v: want exactly my two artifacts", c.ChooseCards)
	}
	if len(c.ChooseSuggested) != 1 || c.ChooseSuggested[0] != rock {
		t.Errorf("suggested %v, want the noncreature artifact", c.ChooseSuggested)
	}
	if err := g.ResolveEntryCardChoice(c.ID, me.ID, nil); err == nil {
		t.Error("an empty answer was accepted: the clause is not a may")
	}

	answerEntryChoice(t, g, game.PendingChoiceEntryChoosePermanents, me.ID, golem)
	if got := b12Counter(t, g, golem, phylacteryCounter); got != 1 {
		t.Errorf("phylactery counters on the golem = %d, want 1", got)
	}
	if got := b12Counter(t, g, rock, phylacteryCounter); got != 0 {
		t.Errorf("phylactery counters on the other artifact = %d, want 0", got)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, id); !ok {
		t.Fatal("the Lich is not on the battlefield after the choice")
	}
}

func TestPhylacteryLichWithNoArtifactEntersAndIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := castPhylacteryLich(t, g)
	if entryChoiceOfKind(g, game.PendingChoiceEntryChoosePermanents, me.ID) != nil {
		t.Fatal("asked a question with no answer")
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, id); ok {
		t.Error("the Lich with no phylactery is still on the battlefield")
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the Lich was not sacrificed to its graveyard")
	}
}

func TestPhylacteryLichIsSacrificedWhenThePhylacteryGoes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := b12Permanent(g, me.ID, "Mind Stone", "Artifact")
	id := castPhylacteryLich(t, g)
	answerEntryChoice(t, g, game.PendingChoiceEntryChoosePermanents, me.ID, rock)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, id); !ok {
		t.Fatal("the Lich did not stay with its phylactery in play")
	}

	b14Kill(g, rock)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, id); ok {
		t.Error("the Lich survived its phylactery (indestructible does not stop a sacrifice)")
	}
}

func TestPhylacteryLichCounterIsDoubledByDoublingSeason(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
	rock := b12Permanent(g, me.ID, "Mind Stone", "Artifact")
	castPhylacteryLich(t, g)
	answerEntryChoice(t, g, game.PendingChoiceEntryChoosePermanents, me.ID, rock)
	if got := b12Counter(t, g, rock, phylacteryCounter); got != 2 {
		t.Errorf("phylactery counters = %d, want 2 (Doubling Season)", got)
	}
}
