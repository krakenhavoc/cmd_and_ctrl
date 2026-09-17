package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestPanharmoniconUsesEnteringTypesUnderMycosynthLattice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Panharmonicon", "Artifact", "76678885-3674-443d-b9a2-2a460cf6aac0", false)
	pushCatalogPermanent(g, me.ID, "Mycosynth Lattice", "Artifact", "ae1f2ab5-c6a5-4d49-a746-3cb4668bf805", false)
	pushCatalogPermanent(g, me.ID, "Ruin Crab", "Creature — Crab", "8afc00d4-a1c6-4329-af2c-a7f58a0c33e7", false)
	before := opp.Library.Size()
	// Use the ordinary land-play path. Lattice makes the Forest enter as
	// an artifact, so Panharmonicon doubles the Crab's landfall trigger.
	castCatalogSpell(t, g, "Forest", "Basic Land — Forest", "", nil)
	passPriorityAroundTable(t, g)
	if got := before - opp.Library.Size(); got != 6 {
		t.Fatalf("artifact landfall milled %d cards, want 6", got)
	}
}

// A linked return sees the cards exiled by BOTH independently targeted
// instances. A per-source single-card slot would lose the first one.
func TestDoubledAngelOfSerenityReturnsEveryLinkedCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Panharmonicon", "Artifact", "76678885-3674-443d-b9a2-2a460cf6aac0", false)
	a := b12Creature(g, opp.ID, "first exile", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "second exile", "Creature — Bear", 3, 3)
	angel := castCatalogSpell(t, g, "Angel of Serenity", "Creature — Angel", b36AngelOfSerenityOracle, nil)
	passPriorityAroundTable(t, g)
	var prompts []uuid.UUID
	for _, ch := range g.PendingChoices {
		if ch.Kind == game.PendingChoiceTriggerPrompt && ch.Source == angel {
			prompts = append(prompts, ch.ID)
		}
	}
	if len(prompts) != 2 {
		t.Fatalf("optional prompts = %d, want 2", len(prompts))
	}
	for _, id := range prompts {
		if err := g.ResolveTriggerPrompt(id, me.ID, true); err != nil {
			t.Fatal(err)
		}
	}
	var picks []uuid.UUID
	for _, ch := range g.PendingChoices {
		if ch.Kind == game.PendingChoicePickTarget && ch.Source == angel {
			picks = append(picks, ch.ID)
		}
	}
	if len(picks) != 2 {
		t.Fatalf("target prompts = %d, want 2", len(picks))
	}
	for i, id := range picks {
		if err := g.ResolvePickTargets(id, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: []uuid.UUID{a, b}[i]}}); err != nil {
			t.Fatal(err)
		}
	}
	passPriorityAroundTable(t, g)
	if !exileHas(g, a) || !exileHas(g, b) {
		t.Fatal("each trigger must exile its own target")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(angel); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(a) || !opp.Hand.Contains(b) {
		t.Fatal("linked return lost one doubled instance's exile")
	}
}

func TestPanharmoniconDoesNotDoubleSagaChapterOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Panharmonicon", "Artifact", "76678885-3674-443d-b9a2-2a460cf6aac0", false)
	// Give the Saga the artifact type so the entry predicate could match.
	// CR 714.2b and 714.3a (pinned August 2026 text): its chapter is
	// caused by the lore counter, not by entering as an artifact.
	id := castCatalogSpell(t, g, "History of Benalia", "Artifact Enchantment — Saga", "c15bb7eb-aaaa-4468-9641-8f706d6137e8", nil)
	passPriorityAroundTable(t, g)
	knights := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Knight" {
			knights++
		}
	}
	if knights != 1 || counterCount(g, id, "lore") != 1 {
		t.Fatalf("entry made %d Knights with %d lore counters, want 1 each", knights, counterCount(g, id, "lore"))
	}
}
