package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// face_down_listed_test.go — the proof cards for CR 708.2's LISTED
// characteristics (#1270, ADR 0082's second 2026-09-23 amendment):
// Cyber Conversion (the turn), Yedora, Grave Gardener (the graveyard
// return, and a body that is not a creature) and Cybership (the
// library put).

const (
	yedoraOracle    = "1fa20b05-cc06-4fb7-ae76-677a29e2185a"
	cybershipOracle = "9c6ac895-0644-47ea-8a40-8f35cbbeba42"
)

// TestCyberConversionMakesACybermanArtifactCreature is the caveat
// #1209 declared, closed: the second sentence's types land.
func TestCyberConversionMakesACybermanArtifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := effectstest.PushCreatureToBattlefieldForTest(g, g.Seats[1].ID, "Sheoldred")

	effectstest.CastCatalogSpell(t, g, "Cyber Conversion", "Instant",
		"33761c1a-9848-45bf-b934-123cebab566b",
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.IsCyberman(t, g, victim)
	c, _ := effectstest.BattlefieldCard(g, victim)
	if c.HasSubtype("Test") {
		t.Error("the card underneath's creature type survived CR 708.2")
	}
}

// TestYedoraReturnsTheDeadAsAFaceDownForest: another nontoken creature
// you control dies; you say yes; it comes back face down as a Forest
// LAND — not a creature — under its owner's control, and it taps for
// {G}.
func TestYedoraReturnsTheDeadAsAFaceDownForest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Yedora, Grave Gardener", "Legendary Creature — Treefolk Druid", yedoraOracle, 5, 5)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	c, ok := effectstest.BattlefieldCard(g, bear)
	if !ok {
		t.Fatal("the Bear did not come back")
	}
	if !c.FaceDown || c.FaceDownKind != game.FaceDownTurned {
		t.Fatalf("state = (%v, %q), want face down, turned", c.FaceDown, c.FaceDownKind)
	}
	if c.IsCreature() || !c.IsLand() || !c.HasSubtype("Forest") {
		t.Errorf("creature %v, land %v, Forest %v — want a Forest land and not a creature",
			c.IsCreature(), c.IsLand(), c.HasSubtype("Forest"))
	}
	if c.Controller != me.ID || !c.IsKnownTo(me.ID) || c.IsKnownTo(opp.ID) {
		t.Error("CR 708.5: its controller, and nobody else, may look at it")
	}
	var abs []game.ManaAbilityShape
	g.ReadSnapshot(func() { abs = game.ManaAbilitiesForCard(c) })
	if len(abs) != 1 || abs[0].Produced != "{G}" {
		t.Errorf("mana abilities = %+v, want CR 305.6's {T}: Add {G}", abs)
	}
}

// TestYedoraIgnoresTokensAndOpponentsAndMayBeDeclined: "another
// NONTOKEN creature YOU CONTROL", and "you MAY".
func TestYedoraIgnoresTokensAndOpponentsAndMayBeDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Yedora, Grave Gardener", "Legendary Creature — Treefolk Druid", yedoraOracle, 5, 5)
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	mine := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == game.PendingChoiceTriggerPrompt {
			t.Fatal("an opponent's creature dying asked Yedora's question")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(mine) {
		t.Error("declined, the Bear should have stayed in the graveyard")
	}
}

// TestCybershipStealsTheTopTwoAsCybermen: combat damage to a player
// puts the top two cards of THAT player's library onto the battlefield
// face down under the Cybership's controller — as 2/2 Cyberman
// artifact creatures, whatever the cards are, an instant included.
func TestCybershipStealsTheTopTwoAsCybermen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Crewed: an artifact creature this turn, which is what lets it
	// deal combat damage at all. Crew is a separate ability and not
	// what this test is about.
	ship := effectstest.B12Push(g, me.ID, "Cybership", "Artifact Creature — Vehicle", cybershipOracle, 8, 8)
	effectstest.AdvanceToMain(t, g)

	bolt := game.Card{InstanceID: uuid.New(), Name: "Lightning Bolt", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID}
	angel := game.Card{InstanceID: uuid.New(), Name: "Serra Angel", TypeLine: "Creature — Angel", Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID}
	third := game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID}
	g.WithWriteLock(func() {
		opp.Library.PushTop(third)
		opp.Library.PushTop(angel)
		opp.Library.PushTop(bolt)
	})

	effectstest.DealCombatDamageToPlayer(g, ship, opp.ID, 8)
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{bolt.InstanceID, angel.InstanceID} {
		effectstest.IsCyberman(t, g, id)
		c, _ := effectstest.BattlefieldCard(g, id)
		if c.Controller != me.ID || c.Owner != opp.ID {
			t.Errorf("controller/owner = %s/%s, want the Cybership's controller / the library's owner", c.Controller, c.Owner)
		}
		if !c.IsKnownTo(me.ID) || c.IsKnownTo(opp.ID) {
			t.Error("CR 708.5: only the controller may look — the owner included")
		}
	}
	if !opp.Library.Contains(third.InstanceID) {
		t.Error("the third card moved — Cybership takes the top TWO")
	}
}
