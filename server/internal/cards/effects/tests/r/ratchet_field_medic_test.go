package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	cardsr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/r"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func ratchetRow() cards.Card {
	return effectstest.TransformerRow(cardsr.RatchetOracleID,
		effectstest.DisturbFace{Name: "Ratchet, Field Medic", TypeLine: "Legendary Artifact Creature — Robot", Cost: "{2}{W}", Power: "2", Toughness: "4"},
		effectstest.DisturbFace{Name: "Ratchet, Rescue Racer", TypeLine: "Legendary Artifact — Vehicle", Cost: "", Power: "1", Toughness: "4"},
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

// "Whenever you gain life, you may convert Ratchet. When you do, return
// target artifact card with mana value at most the life you gained this
// turn": converts, then returns the artifact tapped.
func TestRatchetConvertsAndReturnsAnArtifactTapped(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	id := effectstest.TransformerIn(t, g, ratchetRow(), me)
	cheap := artifactInGraveyard(me, "Cheap Rock", "{2}")
	artifactInGraveyard(me, "Dear Rock", "{5}")

	effectstest.GainLifeNow(t, g, id, me, 3)
	effectstest.TfSettle(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PickCard(t, g, me.ID, cheap)
	effectstest.TfSettle(t, g)

	c := effectstest.WantFace(t, g, id, 1, "Ratchet, Rescue Racer")
	if c.TurnOvers != 1 {
		t.Errorf("turn-overs = %d, want 1", c.TurnOvers)
	}
	back := effectstest.FlipState(t, g, cheap)
	if !back.Tapped {
		t.Error("the returned artifact entered untapped")
	}
	if zone, _ := effectstest.CardWhere(g, me, cheap); zone != "battlefield" {
		t.Errorf("the artifact is in %s, want the battlefield", zone)
	}
}

// The artifact must cost no more than the life gained THIS turn: a
// five-drop is not a target after three life.
func TestRatchetCannotReturnAnArtifactCostingMoreThanTheLifeGained(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	id := effectstest.TransformerIn(t, g, ratchetRow(), me)
	dear := artifactInGraveyard(me, "Dear Rock", "{5}")

	effectstest.GainLifeNow(t, g, id, me, 3)
	effectstest.TfSettle(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.TfSettle(t, g)
	effectstest.WantFace(t, g, id, 1, "Ratchet, Rescue Racer")
	if zone, _ := effectstest.CardWhere(g, me, dear); zone != "graveyard" {
		t.Errorf("the {5} artifact is in %s, want it left in the graveyard", zone)
	}
}

// Declining the "may" converts nothing and returns nothing.
func TestRatchetDeclinedDoesNothing(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	id := effectstest.TransformerIn(t, g, ratchetRow(), me)
	cheap := artifactInGraveyard(me, "Cheap Rock", "{1}")
	effectstest.GainLifeNow(t, g, id, me, 3)
	effectstest.TfSettle(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.TfSettle(t, g)
	effectstest.WantFace(t, g, id, 0, "Ratchet, Field Medic")
	if zone, _ := effectstest.CardWhere(g, me, cheap); zone != "graveyard" {
		t.Errorf("the artifact is in %s, want the graveyard", zone)
	}
}

// CR 701.28e: if Ratchet turned over while the trigger was waiting, the
// convert is ignored, so "when you do" never happens.
func TestRatchetAlreadyTurnedOverReturnsNothing(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	id := effectstest.TransformerIn(t, g, ratchetRow(), me)
	cheap := artifactInGraveyard(me, "Cheap Rock", "{1}")
	effectstest.GainLifeNow(t, g, id, me, 3)
	g.WithWriteLock(func() {
		if err := g.ConvertPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.TfSettle(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.TfSettle(t, g)
	c := effectstest.FlipState(t, g, id)
	if c.TurnOvers != 1 {
		t.Errorf("turn-overs = %d, want the trigger's convert ignored", c.TurnOvers)
	}
	if zone, _ := effectstest.CardWhere(g, me, cheap); zone != "graveyard" {
		t.Errorf("the artifact is in %s, want the graveyard (nothing was converted)", zone)
	}
}

// Rescue Racer converts back when a nontoken artifact you control goes
// to a graveyard, and not for a token.
func TestRatchetRescueRacerConvertsOnANontokenArtifactLeaving(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	id := effectstest.TransformerConverted(t, g, ratchetRow(), me)

	tok := effectstest.PushCatalogPermanent(g, me.ID, "Servo", "Token Artifact Creature — Servo", "", false)
	effectstest.B25Destroy(g, tok)
	effectstest.TfSettle(t, g)
	effectstest.WantFace(t, g, id, 1, "Ratchet, Rescue Racer")

	rock := effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	effectstest.B25Destroy(g, rock)
	effectstest.TfSettle(t, g)
	effectstest.WantFace(t, g, id, 0, "Ratchet, Field Medic")
}

// "This ability triggers only once each turn", and several artifacts
// dying together are one trigger.
func TestRatchetRescueRacerTriggersOnlyOncePerTurn(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	id := effectstest.TransformerConverted(t, g, ratchetRow(), me)

	one := effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	two := effectstest.PushCatalogPermanent(g, me.ID, "Stone", "Artifact", "", false)
	effectstest.B25Destroy(g, one)
	effectstest.B25Destroy(g, two)
	effectstest.TfSettle(t, g)
	c := effectstest.WantFace(t, g, id, 0, "Ratchet, Field Medic")
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
	three := effectstest.PushCatalogPermanent(g, me.ID, "Pebble", "Artifact", "", false)
	effectstest.B25Destroy(g, three)
	effectstest.TfSettle(t, g)
	effectstest.WantFace(t, g, id, 1, "Ratchet, Rescue Racer")
}

// Lifelink damage and Ratchet's death can come together: a blocked
// Ratchet gains its controller life and dies in the same combat damage
// step. Its "you may convert Ratchet" then resolves with Ratchet in the
// graveyard. CR 701.28a converts a permanent, so nothing happens: no
// convert, no "when you do", and no effect error (catalog soak seed
// 2026101600).
func TestRatchetGoneBeforeItsLifeTriggerResolvesConvertsNothing(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	id := effectstest.TransformerIn(t, g, ratchetRow(), me)
	cheap := artifactInGraveyard(me, "Cheap Rock", "{2}")

	effectstest.GainLifeNow(t, g, id, me, 3)
	effectstest.B25Destroy(g, id)
	effectstest.TfSettle(t, g)
	if c := effectstest.LatestConfirmFor(g, me.ID); c != nil {
		if err := g.ResolveConfirm(c.ID, me.ID, true); err != nil {
			t.Fatalf("ResolveConfirm: %v", err)
		}
		effectstest.TfSettle(t, g)
	}
	if errs := effectstest.EffectErrors(g); len(errs) != 0 {
		t.Fatalf("effect errors: %v", errs)
	}
	if zone, _ := effectstest.CardWhere(g, me, cheap); zone != "graveyard" {
		t.Errorf("the artifact is in %s, want it left in the graveyard", zone)
	}
}

// Rescue Racer is itself a nontoken artifact: dying beside another one
// triggers it, and the convert finds no permanent to turn over.
func TestRatchetRescueRacerDyingWithAnotherArtifactConvertsNothing(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	id := effectstest.TransformerConverted(t, g, ratchetRow(), me)
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)

	effectstest.B25Destroy(g, rock)
	effectstest.B25Destroy(g, id)
	effectstest.TfSettle(t, g)
	if errs := effectstest.EffectErrors(g); len(errs) != 0 {
		t.Fatalf("effect errors: %v", errs)
	}
}
