package aiseat_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// autopay_rider_bot_test.go — #3853. A bot pays its casts with the same
// auto-tap planner a human's click does: the enumerator marks a cast it
// can afford AutoTap, and the engine plans the payment when the move is
// dispatched. So a card that reads its payment ("if {R} was spent to cast
// it") gets the red mana from a bot's cast too.
//
// Not behind AISEAT_GAME_TESTS: one enumeration, one dispatch.

const (
	botSolRingOracle       = "6ad8011d-3471-4369-9d68-b264cc027487"
	botGruulScrapperOracle = "878e8980-a2cc-4335-aafe-f5166ef48f79"
)

func TestABotsAutoPaidCastSpendsTheColourTheCardReads(t *testing.T) {
	g := newSettledTable(t, 3853)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToStep(t, g, game.StepPrecombatMain)
	me.Hand.Cards = nil

	// The Swamp arrives before the Mountain and the Sol Ring is beside
	// them, so the colourless-first instinct would pay the generic half
	// with the Sol Ring and a Swamp and never spend a red mana.
	for _, land := range []struct{ name, typeLine string }{
		{"Forest", "Basic Land — Forest"},
		{"Swamp", "Basic Land — Swamp"},
		{"Mountain", "Basic Land — Mountain"},
	} {
		if _, err := g.SpawnCards(me.ID, me.ID, game.ZoneBattlefield, game.Card{Name: land.name, TypeLine: land.typeLine}, 1); err != nil {
			t.Fatalf("spawn %s: %v", land.name, err)
		}
	}
	if _, err := g.SpawnCards(me.ID, me.ID, game.ZoneBattlefield, game.Card{
		Name: "Sol Ring", TypeLine: "Artifact", OracleID: botSolRingOracle,
	}, 1); err != nil {
		t.Fatalf("spawn Sol Ring: %v", err)
	}

	scrapper := uuid.New()
	card := game.Card{
		InstanceID: scrapper, Name: "Gruul Scrapper", OracleID: botGruulScrapperOracle,
		TypeLine: "Creature — Human Berserker", ManaCost: "{3}{G}", Power: 3, Toughness: 2,
		Colors: []string{"G"}, Owner: me.ID, Controller: me.ID,
	}
	card.AddKnowersAll(seatIDs(g))
	me.Hand.PushTop(card)

	var cast *legal.Move
	moves := legal.EnumerateFor(g, me.ID)
	for i := range moves {
		if moves[i].Type == legal.TypeCastSpell && moves[i].Source == scrapper {
			cast = &moves[i]
			break
		}
	}
	if cast == nil {
		t.Fatalf("the enumerator offers no cast of the Scrapper; moves: %v", attackLabels(moves))
	}
	var params struct {
		AutoTap bool `json:"auto_tap"`
	}
	if err := json.Unmarshal(cast.Params, &params); err != nil || !params.AutoTap {
		t.Fatalf("the cast move does not ask for auto-pay: %s (%v)", cast.Params, err)
	}

	if err := actions.Dispatch(g, actions.Action{
		Type: actions.Type(cast.Type), Player: cast.Player, Caller: me.ID, Params: cast.Params,
	}); err != nil {
		t.Fatalf("the engine refused the enumerated cast %q: %v", cast.Label, err)
	}
	item := g.StackMeta[scrapper]
	if item == nil {
		t.Fatal("the Scrapper is not on the stack")
	}
	if item.Paid.SpentOfColor("R") != 1 {
		t.Errorf("{R} spent = %d, want 1 — the bot's auto-pay should reach the Mountain (%+v)",
			item.Paid.SpentOfColor("R"), item.Paid.Mana)
	}
}
