package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// untap_hold_cards_test.go — #1313's proof cards: untap holds with a
// CR 611.2b "for as long as" duration (ADR 0058's 2026-09-23
// amendment).

const (
	OracleTyLee          = "081ad4e3-cda3-41cc-890f-412611dc9ea0"
	OracleDungeonGeists  = "ab5ebae2-cd77-4a7d-a93b-8042cd486429"
	OracleTidebinderMage = "f881378b-b539-4ea8-981d-e01ee82af105"
)

// castHoldCreature casts a creature spell with real P/T from the active
// player's hand and resolves it, leaving its ETB trigger waiting on its
// target prompt. castCatalogSpell's P/T-less creature would die to
// CR 704.5f before the "for as long as you control ~" hold could start.
func CastHoldCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	PassPriorityAroundTable(t, g)
	return id
}

// pickTargetPrompt returns the open pick_target prompt, failing the
// test if there is none.
func PickTargetPrompt(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			return c
		}
	}
	t.Fatal("no pick_target prompt is open")
	return nil
}

func AnswerPickTarget(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	ch := PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: target}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

func HoldCount(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("battlefield card %s is missing", id)
	}
	n := 0
	for _, skip := range c.NextUntapSkips {
		if skip.While != nil {
			n++
		}
	}
	return n
}
