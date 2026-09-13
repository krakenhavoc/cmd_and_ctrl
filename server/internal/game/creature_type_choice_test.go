package game

import (
	"testing"

	"github.com/google/uuid"
)

// creature_type_choice_test.go pins the CR 614.12 prompt: who may
// answer it, what counts as an answer, and the invalidation that
// makes the answer visible to the layer engine.

func queueTypeChoice(t *testing.T, g *Game, chooser, source uuid.UUID) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueCreatureTypeChoiceForEffect(chooser, source, "Cavern of Souls")
	})
	return id
}

func TestResolveCreatureTypeChoiceStampsNamedTribe(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	cavern := pushTypedTestCard(g, Card{
		Name: "Cavern of Souls", TypeLine: "Land", Owner: seat, Controller: seat,
	})
	choiceID := queueTypeChoice(t, g, seat, cavern)

	// Canonicalisation: a lower-case answer is accepted and stored in
	// the vocabulary's spelling.
	if err := g.ResolveCreatureTypeChoice(choiceID, seat, "elf"); err != nil {
		t.Fatalf("ResolveCreatureTypeChoice: %v", err)
	}
	card := layeredBattlefieldCard(t, g, cavern)
	if card.NamedTribe != "Elf" {
		t.Errorf("NamedTribe = %q, want %q", card.NamedTribe, "Elf")
	}
	var pending int
	g.ReadSnapshot(func() { pending = len(g.PendingChoices) })
	if pending != 0 {
		t.Errorf("pending choices after answering = %d, want 0", pending)
	}
}

func TestResolveCreatureTypeChoiceRejectsJunk(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	other := g.Seats[1].ID
	cavern := pushTypedTestCard(g, Card{
		Name: "Cavern of Souls", TypeLine: "Land", Owner: seat, Controller: seat,
	})
	choiceID := queueTypeChoice(t, g, seat, cavern)

	// Not a creature type at all.
	if err := g.ResolveCreatureTypeChoice(choiceID, seat, "Equipment"); err == nil {
		t.Error(`naming "Equipment" was accepted, want ErrInvalidParam`)
	}
	// A land type is not a creature type either.
	if err := g.ResolveCreatureTypeChoice(choiceID, seat, "Forest"); err == nil {
		t.Error(`naming "Forest" was accepted, want ErrInvalidParam`)
	}
	// Someone else's prompt.
	if err := g.ResolveCreatureTypeChoice(choiceID, other, "Elf"); err == nil {
		t.Error("a non-chooser answered the prompt, want ErrNotTheChooser")
	}
	// The prompt survives every rejection — a bad answer must not
	// consume the choice and strand the permanent with no tribe.
	var pending int
	g.ReadSnapshot(func() { pending = len(g.PendingChoices) })
	if pending != 1 {
		t.Fatalf("pending choices after three rejected answers = %d, want 1", pending)
	}
	if err := g.ResolveCreatureTypeChoice(choiceID, seat, "Sliver"); err != nil {
		t.Fatalf("valid answer after rejections: %v", err)
	}
}

// TestResolveCreatureTypeChoiceInvalidatesLayerCache — the named
// tribe is an AppliesTo input, and nothing else on this path emits an
// event the layer listener watches. Without the explicit bump a
// lord's +1/+1 would appear only when some unrelated permanent moved.
func TestResolveCreatureTypeChoiceInvalidatesLayerCache(t *testing.T) {
	g := newActiveGame(t)
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != "named-tribe-anthem" {
			return nil
		}
		return []StaticAbility{{
			Layer:    Layer7PT,
			SubLayer: SubLayer7C_Modify,
			AppliesTo: func(target *Card, g *Game, source *Card) bool {
				return source.NamedTribe != "" && target.IsCreature() &&
					target.HasSubtype(source.NamedTribe)
			},
			Apply: func(c *Characteristic, target *Card, g *Game, source *Card) {
				c.Power++
				c.Toughness++
			},
		}}
	})
	seat := g.Seats[0].ID
	elf := pushTypedTestCard(g, Card{
		Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid",
		Power: 1, Toughness: 1, Owner: seat, Controller: seat,
	})
	banner := pushTypedTestCard(g, Card{
		Name: "Vanquisher's Banner", TypeLine: "Artifact",
		OracleID: "named-tribe-anthem", Owner: seat, Controller: seat,
	})

	// Force a recompute so the cache is warm and stale-looking.
	if got := layeredBattlefieldCard(t, g, elf).CurrentPower(); got != 1 {
		t.Fatalf("before the choice: power = %d, want 1", got)
	}

	choiceID := queueTypeChoice(t, g, seat, banner)
	if err := g.ResolveCreatureTypeChoice(choiceID, seat, "Elf"); err != nil {
		t.Fatalf("ResolveCreatureTypeChoice: %v", err)
	}
	if got := layeredBattlefieldCard(t, g, elf).CurrentPower(); got != 2 {
		t.Errorf("after naming Elf: power = %d, want 2 (the layer cache did not invalidate)", got)
	}
}

// TestResolveCreatureTypeChoiceSurvivesASourceThatLeft — the prompt
// is asynchronous, so the permanent can be gone when the answer
// lands. That is not an error: the choice is made and has nowhere to
// go, which is what CR 608.2 does with a missing object.
func TestResolveCreatureTypeChoiceSurvivesASourceThatLeft(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	cavern := pushTypedTestCard(g, Card{
		Name: "Cavern of Souls", TypeLine: "Land", Owner: seat, Controller: seat,
	})
	choiceID := queueTypeChoice(t, g, seat, cavern)
	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, g.Seats[0].Graveyard, cavern); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
	})
	if err := g.ResolveCreatureTypeChoice(choiceID, seat, "Elf"); err != nil {
		t.Fatalf("answering for a permanent that left: %v", err)
	}
}
