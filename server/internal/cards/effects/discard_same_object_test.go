package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// discard_same_object_test.go — #2453: a discard trigger that exiles
// "that card" must act on the object that was discarded. A card that
// left the graveyard and came back before the trigger resolved is a new
// object (CR 400.7) and stays where it is.

const containmentConstructOracle = "54d48e54-be4a-4a78-a778-b28e74ef7134"

type discardExileCase struct {
	name string
	// setup puts the source on seat 0's battlefield and returns the seat
	// that discards.
	setup func(g *game.Game) (victim *game.Player)
	// optional answers a "you may" prompt, when the card asks one.
	optional bool
}

func discardExileCases() []discardExileCase {
	return []discardExileCase{
		{name: "Tinybones", setup: func(g *game.Game) *game.Player {
			pushCatalogPermanent(g, g.Seats[0].ID, "Tinybones, Bauble Burglar", "Legendary Creature — Skeleton Rogue", tinybonesOracle, false)
			return g.Seats[1]
		}},
		{name: "Necropotence", setup: func(g *game.Game) *game.Player {
			seedReplacementPermanent(g, necropotenceOracle, "Necropotence", g.Seats[0].ID)
			return g.Seats[0]
		}},
		{name: "Bag of Holding", setup: func(g *game.Game) *game.Player {
			b12Push(g, g.Seats[0].ID, "Bag of Holding", "Artifact", bagOfHoldingOracle, 0, 0)
			return g.Seats[0]
		}},
		{name: "Containment Construct", optional: true, setup: func(g *game.Game) *game.Player {
			b12Push(g, g.Seats[0].ID, "Containment Construct", "Artifact Creature — Construct", containmentConstructOracle, 2, 1)
			return g.Seats[0]
		}},
	}
}

func runDiscardExile(t *testing.T, tc discardExileCase, cycle bool) (g *game.Game, victim *game.Player, card uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	advanceToMain(t, g)
	victim = tc.setup(g)
	card = soleHandCardForTest(victim, "Chaff")
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(victim.ID, 1); err != nil {
			t.Fatalf("discard: %v", err)
		}
	})
	if !victim.Graveyard.Contains(card) {
		t.Fatal("the discard should land in the graveyard first")
	}
	if cycle {
		// In response: the card goes back to hand and is binned again
		// without a discard, so it is a new object with no new trigger.
		gy := game.ZoneRef{Kind: game.ZoneGraveyard, Owner: victim.ID}
		hand := game.ZoneRef{Kind: game.ZoneHand, Owner: victim.ID}
		if err := g.MoveCardByID(gy, hand, card); err != nil {
			t.Fatal(err)
		}
		if err := g.MoveCardByID(hand, gy, card); err != nil {
			t.Fatal(err)
		}
	}
	if tc.optional {
		answerLatestTriggerPrompt(t, g, g.Seats[0].ID, true)
	}
	passPriorityAroundTable(t, g)
	return g, victim, card
}

func TestDiscardExileTriggersExileTheDiscardedCard(t *testing.T) {
	for _, tc := range discardExileCases() {
		t.Run(tc.name, func(t *testing.T) {
			g, victim, card := runDiscardExile(t, tc, false)
			if victim.Graveyard.Contains(card) || !g.Exile.Contains(card) {
				t.Fatal("the discarded card should be exiled when it never left the graveyard")
			}
		})
	}
}

func TestDiscardExileTriggersLeaveANewObjectAlone(t *testing.T) {
	for _, tc := range discardExileCases() {
		t.Run(tc.name, func(t *testing.T) {
			g, victim, card := runDiscardExile(t, tc, true)
			if g.Exile.Contains(card) || !victim.Graveyard.Contains(card) {
				t.Fatal("a card that left the graveyard and came back is a new object and must not be exiled (CR 400.7)")
			}
		})
	}
}
