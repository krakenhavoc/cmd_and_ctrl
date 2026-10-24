package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// rebound_cast_moves_test.go — #1854: the bot half of rebound.
//
// Rebound adds no enumerator code, which is the claim under test. The
// upkeep offer is a resolution_cast prompt (#4250, resolution_cast.go),
// and the cast is an ordinary granted cast out of exile made during the
// trigger's resolution. What a bot must not do is stall: it is offered
// the free cast of a SORCERY in the upkeep, because the grant carries
// CR 608.2g's timing, and the always-legal decline, and nothing else.
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
			if c != nil && c.Kind == game.PendingChoiceResolutionCast {
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
	var casts []legal.Move
	declines := 0
	for _, m := range moves {
		switch {
		case m.ResolutionCast != offer.ID.String():
			t.Errorf("a seat owing the rebound cast was offered %q as well", m.Label)
		case m.Type == legal.TypeResolveChoice && m.AlwaysLegal:
			declines++
		case m.Kind == legal.KindCast && m.Source == id:
			casts = append(casts, m)
		}
	}
	if len(casts) != 1 || declines != 1 {
		t.Fatalf("enumerated %v, want the one free cast and the decline", labels(moves))
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
	// The move is accepted as written: the cast answers the prompt.
	m := casts[0]
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: me.ID, Params: m.Params}); err != nil {
		t.Fatalf("the enumerated cast was refused: %v", err)
	}
	if !g.Stack.Contains(id) || g.ResolutionPaused() {
		t.Error("the rebound cast is not on the stack, or the trigger is still resolving")
	}
}
