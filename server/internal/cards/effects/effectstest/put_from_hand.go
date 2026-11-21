package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// put_from_hand_test.go — the catalog half of #654: the shared
// MayPutALandFromHand clause and the five cards that print it.
//
// The engine half (the move itself, the CR 614 pipeline, the land
// drop that isn't one) is pinned in
// server/internal/game/put_from_hand_test.go. What these tests are
// about is the PROMPT: who is offered what, what a decline does, and
// that the card that arrives went through a real entry.

const (
	PutFromHandEurekaOracle     = "0e2c11b2-d95f-4402-9a4a-afd3f7ffb8be"
	PutFromHandBrokenBondOracle = "858e12e9-3eaa-40cf-9e22-f9ccdfe485b3"
	PutFromHandFungusOracle     = "0a8d0217-ff24-4177-b6be-707eb2b6b9e9"
	PutFromHandSpelunkingOracle = "2962fe4c-bf48-454b-8a6b-0f8253352ae8"
	PutFromHandKhalniOracle     = "b2d5ba45-8674-4428-89db-c2bbbf0bf5c5"
)

// handCardForTest seeds one card into a player's hand.
func HandCardForTest(p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// answerChooseCards answers the open card-set prompt for a player.
func AnswerChooseCards(t *testing.T, g *game.Game, chooser uuid.UUID, picks ...uuid.UUID) {
	t.Helper()
	pick := ChooseCardsChoiceFor(g, chooser)
	if pick == nil {
		t.Fatal("no choose-cards prompt is open")
	}
	if err := g.ResolveChooseCards(pick.ID, chooser, picks); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
}

// --- small local helpers -------------------------------------------

func TappedOnBattlefield(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Tapped
		}
	}
	t.Fatalf("card %s is not on the battlefield", id)
	return false
}
