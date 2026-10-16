package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func ratchetRow() cards.Card {
	return transformerRow(ratchetOracleID,
		disturbFace{"Ratchet, Field Medic", "Legendary Artifact Creature — Robot", "{2}{W}", "2", "4"},
		disturbFace{"Ratchet, Rescue Racer", "Legendary Artifact — Vehicle", "", "1", "4"},
		"More Than Meets the Eye {1}{W}\nLifelink\nWhenever you gain life, you may convert Ratchet. When you do, return target artifact card with mana value less than or equal to the amount of life you gained this turn from your graveyard to the battlefield tapped.",
		"Living metal\nLifelink\nWhenever one or more nontoken artifacts you control are put into a graveyard from the battlefield, convert Ratchet. This ability triggers only once each turn.",
		"Living metal", "Convert", "Lifelink")
}

// artifactInGraveyard puts an artifact card with the given cost into
// the player's graveyard.
func artifactInGraveyard(p *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Artifact", ManaCost: cost, Owner: p.ID, Controller: p.ID})
	return id
}

func gainLifeNow(t *testing.T, g *game.Game, source uuid.UUID, p *game.Player, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(source, p.ID, n); err != nil {
			t.Fatal(err)
		}
	})
}

// "Whenever you gain life, you may convert Ratchet. When you do, return
// target artifact card with mana value at most the life you gained this
// turn": converts, then returns the artifact tapped.
func TestRatchetConvertsAndReturnsAnArtifactTapped(t *testing.T) {
	g, me, _ := flipMain(t)
	id := transformerIn(t, g, ratchetRow(), me)
	cheap := artifactInGraveyard(me, "Cheap Rock", "{2}")
	artifactInGraveyard(me, "Dear Rock", "{5}")

	gainLifeNow(t, g, id, me, 3)
	tfSettle(t, g)
	answerMayChoice(t, g, me.ID, true)
	pickCard(t, g, me.ID, cheap)
	tfSettle(t, g)

	c := wantFace(t, g, id, 1, "Ratchet, Rescue Racer")
	if c.TurnOvers != 1 {
		t.Errorf("turn-overs = %d, want 1", c.TurnOvers)
	}
	back := flipState(t, g, cheap)
	if !back.Tapped {
		t.Error("the returned artifact entered untapped")
	}
	if zone, _ := cardWhere(g, me, cheap); zone != "battlefield" {
		t.Errorf("the artifact is in %s, want the battlefield", zone)
	}
}

// The artifact must cost no more than the life gained THIS turn: a
// five-drop is not a target after three life.
func TestRatchetCannotReturnAnArtifactCostingMoreThanTheLifeGained(t *testing.T) {
	g, me, _ := flipMain(t)
	id := transformerIn(t, g, ratchetRow(), me)
	dear := artifactInGraveyard(me, "Dear Rock", "{5}")

	gainLifeNow(t, g, id, me, 3)
	tfSettle(t, g)
	answerMayChoice(t, g, me.ID, true)
	tfSettle(t, g)
	wantFace(t, g, id, 1, "Ratchet, Rescue Racer")
	if zone, _ := cardWhere(g, me, dear); zone != "graveyard" {
		t.Errorf("the {5} artifact is in %s, want it left in the graveyard", zone)
	}
}

// Declining the "may" converts nothing and returns nothing.
func TestRatchetDeclinedDoesNothing(t *testing.T) {
	g, me, _ := flipMain(t)
	id := transformerIn(t, g, ratchetRow(), me)
	cheap := artifactInGraveyard(me, "Cheap Rock", "{1}")
	gainLifeNow(t, g, id, me, 3)
	tfSettle(t, g)
	answerMayChoice(t, g, me.ID, false)
	tfSettle(t, g)
	wantFace(t, g, id, 0, "Ratchet, Field Medic")
	if zone, _ := cardWhere(g, me, cheap); zone != "graveyard" {
		t.Errorf("the artifact is in %s, want the graveyard", zone)
	}
}

// CR 701.28e: if Ratchet turned over while the trigger was waiting, the
// convert is ignored, so "when you do" never happens.
func TestRatchetAlreadyTurnedOverReturnsNothing(t *testing.T) {
	g, me, _ := flipMain(t)
	id := transformerIn(t, g, ratchetRow(), me)
	cheap := artifactInGraveyard(me, "Cheap Rock", "{1}")
	gainLifeNow(t, g, id, me, 3)
	g.WithWriteLock(func() {
		if err := g.ConvertPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	tfSettle(t, g)
	answerMayChoice(t, g, me.ID, true)
	tfSettle(t, g)
	c := flipState(t, g, id)
	if c.TurnOvers != 1 {
		t.Errorf("turn-overs = %d, want the trigger's convert ignored", c.TurnOvers)
	}
	if zone, _ := cardWhere(g, me, cheap); zone != "graveyard" {
		t.Errorf("the artifact is in %s, want the graveyard (nothing was converted)", zone)
	}
}

// Rescue Racer converts back when a nontoken artifact you control goes
// to a graveyard, and not for a token.
func TestRatchetRescueRacerConvertsOnANontokenArtifactLeaving(t *testing.T) {
	g, me, _ := flipMain(t)
	id := transformerConverted(t, g, ratchetRow(), me)

	tok := pushCatalogPermanent(g, me.ID, "Servo", "Token Artifact Creature — Servo", "", false)
	b25Destroy(g, tok)
	tfSettle(t, g)
	wantFace(t, g, id, 1, "Ratchet, Rescue Racer")

	rock := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	b25Destroy(g, rock)
	tfSettle(t, g)
	wantFace(t, g, id, 0, "Ratchet, Field Medic")
}

// "This ability triggers only once each turn", and several artifacts
// dying together are one trigger.
func TestRatchetRescueRacerTriggersOnlyOncePerTurn(t *testing.T) {
	g, me, _ := flipMain(t)
	id := transformerConverted(t, g, ratchetRow(), me)

	one := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	two := pushCatalogPermanent(g, me.ID, "Stone", "Artifact", "", false)
	b25Destroy(g, one)
	b25Destroy(g, two)
	tfSettle(t, g)
	c := wantFace(t, g, id, 0, "Ratchet, Field Medic")
	if c.TurnOvers != 2 {
		t.Fatalf("two artifacts together turned it over %d times, want once more", c.TurnOvers-1)
	}

	// Back to Rescue Racer by other means; a third artifact this turn
	// finds the ability already used.
	g.WithWriteLock(func() {
		if err := g.ConvertPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	three := pushCatalogPermanent(g, me.ID, "Pebble", "Artifact", "", false)
	b25Destroy(g, three)
	tfSettle(t, g)
	wantFace(t, g, id, 1, "Ratchet, Rescue Racer")
}

// Lifelink damage and Ratchet's death can come together: a blocked
// Ratchet gains its controller life and dies in the same combat damage
// step. Its "you may convert Ratchet" then resolves with Ratchet in the
// graveyard. CR 701.28a converts a permanent, so nothing happens: no
// convert, no "when you do", and no effect error (catalog soak seed
// 2026101600).
func TestRatchetGoneBeforeItsLifeTriggerResolvesConvertsNothing(t *testing.T) {
	g, me, _ := flipMain(t)
	id := transformerIn(t, g, ratchetRow(), me)
	cheap := artifactInGraveyard(me, "Cheap Rock", "{2}")

	gainLifeNow(t, g, id, me, 3)
	b25Destroy(g, id)
	tfSettle(t, g)
	if c := latestConfirmFor(g, me.ID); c != nil {
		if err := g.ResolveConfirm(c.ID, me.ID, true); err != nil {
			t.Fatalf("ResolveConfirm: %v", err)
		}
		tfSettle(t, g)
	}
	if errs := effectErrors(g); len(errs) != 0 {
		t.Fatalf("effect errors: %v", errs)
	}
	if zone, _ := cardWhere(g, me, cheap); zone != "graveyard" {
		t.Errorf("the artifact is in %s, want it left in the graveyard", zone)
	}
}

// Rescue Racer is itself a nontoken artifact: dying beside another one
// triggers it, and the convert finds no permanent to turn over.
func TestRatchetRescueRacerDyingWithAnotherArtifactConvertsNothing(t *testing.T) {
	g, me, _ := flipMain(t)
	id := transformerConverted(t, g, ratchetRow(), me)
	rock := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)

	b25Destroy(g, rock)
	b25Destroy(g, id)
	tfSettle(t, g)
	if errs := effectErrors(g); len(errs) != 0 {
		t.Fatalf("effect errors: %v", errs)
	}
}
