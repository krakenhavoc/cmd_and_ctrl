package legal_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// rebound_cast_moves_test.go — #1854: the bot half of rebound.
//
// Rebound adds no enumerator code, which is the claim under test. The
// upkeep offer is a may_cast prompt (choices.go's PendingChoiceMayCast
// arm), and the cast is an ordinary granted cast out of exile. What a
// bot must not do is stall: it is offered both answers to the prompt,
// and once it takes the offer it is offered the free cast of a SORCERY
// in the upkeep, because the grant carries CR 608.2g's timing.
func TestTheEnumeratorOffersTheReboundCastInTheUpkeep(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	clearHand(me)
	advanceTo(t, g, game.StepPrecombatMain)
	id := handCard(me, game.Card{
		Name: "Test Rebound Sorcery", TypeLine: "Sorcery", ManaCost: "{1}{U}",
		Layout: "normal", Keywords: []string{game.KeywordRebound},
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast from hand: %v", err)
	}
	for i := 0; i < 8 && !g.Exile.Contains(id); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !g.Exile.Contains(id) {
		t.Fatal("setup: the rebound sorcery was not exiled")
	}

	// Round the table to seat 0's next upkeep.
	for i := 0; i < 80; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == 0 {
			break
		}
	}
	if g.Turn.Step != game.StepUpkeep || g.Turn.ActiveSeat != 0 {
		t.Fatal("setup: never reached seat 0's next upkeep")
	}
	var offer *game.PendingChoice
	for i := 0; i < 8 && offer == nil; i++ {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceMayCast {
				offer = c
			}
		}
		if offer == nil {
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
	}
	if offer == nil {
		t.Fatal("no rebound offer at the upkeep")
	}

	moves := legal.EnumerateFor(g, me.ID)
	var accept *legal.Move
	for i, m := range moves {
		if m.Type != legal.TypeResolveChoice {
			t.Errorf("a seat owing may_cast was offered %q as well", m.Label)
		}
		if strings.HasSuffix(m.Label, ": cast it") {
			accept = &moves[i]
		}
	}
	if len(moves) != 2 || accept == nil {
		t.Fatalf("enumerated %v, want the two may_cast answers", labels(moves))
	}
	if err := g.ResolveMayCast(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}

	casts := castMovesFor(legal.EnumerateFor(g, me.ID), id)
	if len(casts) != 1 {
		t.Fatalf("rebound cast moves in the upkeep = %d, want 1", len(casts))
	}
	var params struct {
		FromZone string `json:"from_zone"`
	}
	if err := json.Unmarshal(casts[0].Params, &params); err != nil {
		t.Fatalf("decode params: %v", err)
	}
	if params.FromZone != "exile" {
		t.Errorf("from_zone = %q, want exile", params.FromZone)
	}
}
