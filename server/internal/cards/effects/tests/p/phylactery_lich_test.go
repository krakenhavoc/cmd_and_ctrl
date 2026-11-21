package testsp

import (
	"testing"

	"github.com/google/uuid"

	cardsp "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/p"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestPhylacteryLichPutsTheCounterBeforeItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	golem := effectstest.B12Creature(g, me.ID, "Steel Golem", "Artifact Creature — Golem", 3, 4)
	rock := effectstest.B12Permanent(g, me.ID, "Mind Stone", "Artifact")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Stone", "Artifact")
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")

	id := effectstest.CastPhylacteryLich(t, g)
	c := effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntryChoosePermanents, me.ID)
	if c == nil {
		t.Fatal("the Lich did not ask for an artifact")
	}
	if _, ok := effectstest.BattlefieldCard(g, id); ok {
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

	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntryChoosePermanents, me.ID, golem)
	if got := effectstest.B12Counter(t, g, golem, cardsp.PhylacteryCounter); got != 1 {
		t.Errorf("phylactery counters on the golem = %d, want 1", got)
	}
	if got := effectstest.B12Counter(t, g, rock, cardsp.PhylacteryCounter); got != 0 {
		t.Errorf("phylactery counters on the other artifact = %d, want 0", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, id); !ok {
		t.Fatal("the Lich is not on the battlefield after the choice")
	}
}

func TestPhylacteryLichWithNoArtifactEntersAndIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.CastPhylacteryLich(t, g)
	if effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntryChoosePermanents, me.ID) != nil {
		t.Fatal("asked a question with no answer")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, id); ok {
		t.Error("the Lich with no phylactery is still on the battlefield")
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the Lich was not sacrificed to its graveyard")
	}
}

func TestPhylacteryLichIsSacrificedWhenThePhylacteryGoes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.B12Permanent(g, me.ID, "Mind Stone", "Artifact")
	id := effectstest.CastPhylacteryLich(t, g)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntryChoosePermanents, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, id); !ok {
		t.Fatal("the Lich did not stay with its phylactery in play")
	}

	effectstest.B14Kill(g, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, id); ok {
		t.Error("the Lich survived its phylactery (indestructible does not stop a sacrifice)")
	}
}

func TestPhylacteryLichCounterIsDoubledByDoublingSeason(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	rock := effectstest.B12Permanent(g, me.ID, "Mind Stone", "Artifact")
	effectstest.CastPhylacteryLich(t, g)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntryChoosePermanents, me.ID, rock)
	if got := effectstest.B12Counter(t, g, rock, cardsp.PhylacteryCounter); got != 2 {
		t.Errorf("phylactery counters = %d, want 2 (Doubling Season)", got)
	}
}
