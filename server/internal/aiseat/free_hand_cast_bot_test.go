package aiseat_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// free_hand_cast_bot_test.go — #3270 through the real card, the real
// enumerator and the real heuristic. "You may cast a spell from your hand
// without paying its mana cost" used to grant a later cast the bot never
// took (the pick was priced as a discard).
//
// #4250: the cast is the resolution_cast prompt's answer, inside the
// resolution. The enumerator offers the hand card's FREE cast only (its
// ordinary cast is not an answer), and the always-legal decline; the bot
// takes the free cast, and the engine accepts it.
//
// Not behind AISEAT_GAME_TESTS: one resolution, a few decisions.

const (
	botRishkarsExpertiseOracle = "97407cd0-2bd2-4074-94d3-4ec3d243fa78"
)

func TestHeuristicSeatPicksAndCastsItsFreeSpell(t *testing.T) {
	g := newSettledTable(t, 3270)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToStep(t, g, game.StepPrecombatMain)
	me.Hand.Cards = nil

	push := func(c game.Card) uuid.UUID {
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = me.ID, me.ID
		c.AddKnowersAll(seatIDs(g))
		me.Hand.PushTop(c)
		return c.InstanceID
	}
	expertise := push(game.Card{Name: "Rishkar's Expertise", OracleID: botRishkarsExpertiseOracle,
		TypeLine: "Sorcery", ManaCost: "{4}{G}", Colors: []string{"G"}})
	// An instant, so its ORDINARY cast is open at this timing too, and two
	// Forests to pay for it: the enumerator must still offer only the free
	// cast as the prompt's answer.
	bears := push(game.Card{Name: "Test Instant", TypeLine: "Instant", ManaCost: "{1}{G}",
		Colors: []string{"G"}})
	g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: me.ID, Controller: me.ID})
	g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: me.ID, Controller: me.ID})

	// The sandbox cast skips the mana gate; the free cast is the thing
	// under test.
	if err := g.CastSpell(me.ID, expertise, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Rishkar's Expertise: %v", err)
	}
	for i := 0; i < 16 && !owesChoice(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !owesChoice(g, me.ID) {
		t.Fatal("Rishkar's Expertise did not offer the free cast")
	}
	var offer *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceResolutionCast && c.Chooser == me.ID {
			offer = c
		}
	}
	if offer == nil || offer.MayCastKeyword != game.MayCastKeywordHandCast || !g.ResolutionPaused() {
		t.Fatalf("offer = %+v, want a hand_cast resolution_cast inside the resolution", offer)
	}
	casts, declines := 0, 0
	for _, mv := range legal.EnumerateFor(g, me.ID) {
		if mv.ResolutionCast != offer.ID.String() {
			t.Errorf("move %q does not answer the offer", mv.Label)
			continue
		}
		if mv.Kind == legal.KindChoice && mv.AlwaysLegal {
			declines++
			continue
		}
		var p struct {
			AlternativeCost string `json:"alternative_cost"`
		}
		if err := json.Unmarshal(mv.Params, &p); err != nil || p.AlternativeCost != game.AltCostKeyGrantedCost {
			t.Errorf("cast %q claims %q, want only the free cast (%q)", mv.Label, p.AlternativeCost, game.AltCostKeyGrantedCost)
		}
		casts++
	}
	if casts == 0 || declines != 1 {
		t.Fatalf("offered %d casts and %d declines", casts, declines)
	}

	taken := driveChoices(t, g, heuristic.New(), me.ID, 4)
	if len(taken) != 1 {
		t.Fatalf("answered the offer with %v", taken)
	}
	if me.Hand.Contains(bears) || !g.Stack.Contains(bears) {
		t.Fatal("the bot did not take the free instant")
	}
	if g.ResolutionPaused() {
		t.Error("the resolution is still held after the one free cast")
	}
	for i := 0; i < 16 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !me.Graveyard.Contains(bears) {
		t.Error("the free instant did not resolve")
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Forest" && c.Tapped {
			t.Error("a Forest was tapped: the instant was cast at its printed cost")
		}
	}
}
