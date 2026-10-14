package aiseat_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// paid_cast_or_else_bot_test.go — #3510 through the real card, the real
// enumerator and the real heuristic. Chandra, Torch of Defiance's first +1
// asks a paid_or_else may_cast; the enumerator offers both answers (the
// decline always legal), the heuristic takes the cast, and the next
// decision is the paid cast from exile, which the engine accepts.
//
// Not behind AISEAT_GAME_TESTS: one resolution, a few decisions.

const botChandraTorchOracle = "12cc800c-c7be-4811-8207-fd01a99cc892"

func TestHeuristicSeatAnswersChandraTorchAndCastsTheExiledCard(t *testing.T) {
	g := newSettledTable(t, 3510)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToStep(t, g, game.StepPrecombatMain)
	// Nothing else to spend the mana on: the decision under test is the
	// exiled card against the pass. (The heuristic does not yet rank a
	// pass-closed cast above an ordinary one; see the PR.)
	me.Hand.Cards = nil
	me.Command.Cards = nil

	chandra := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: chandra, Name: "Chandra, Torch of Defiance", OracleID: botChandraTorchOracle,
		TypeLine: "Legendary Planeswalker — Chandra", ManaCost: "{2}{R}{R}", Colors: []string{"R"},
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLoyalty: 4},
	})
	for i := 0; i < 2; i++ {
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
			Owner: me.ID, Controller: me.ID})
	}
	bears := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: bears, Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})

	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		if p.ID != me.ID {
			lives[p.ID] = p.Life
		}
	}

	if err := g.ActivateCatalogAbility(me.ID, chandra, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Chandra's +1: %v", err)
	}
	for i := 0; i < 16 && !owesChoice(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !owesChoice(g, me.ID) {
		t.Fatal("Chandra's +1 asked nothing")
	}

	// The prompt is answerable both ways, and the decline is the
	// always-legal way out.
	moves := legal.EnumerateFor(g, me.ID)
	var accept, decline bool
	for _, mv := range moves {
		if mv.Type != legal.TypeResolveChoice {
			continue
		}
		var p struct {
			Apply *bool `json:"apply"`
		}
		if json.Unmarshal(mv.Params, &p) == nil && p.Apply != nil {
			accept = accept || *p.Apply
			decline = decline || !*p.Apply
		}
	}
	if !accept || !decline {
		t.Fatalf("the paid_or_else prompt offers cast %v and don't %v; moves %v", accept, decline, attackLabels(moves))
	}

	pol := heuristic.New()
	if taken := driveChoices(t, g, pol, me.ID, 4); len(taken) != 1 {
		t.Fatalf("answered the prompt with %v", taken)
	}

	// The next decisions: anything that keeps the window open (a land
	// drop) is fine, but the bot must cast the Bears before it passes,
	// because its pass is the "don't".
	cast := false
	for step := 0; step < 6 && !cast; step++ {
		moves = legal.EnumerateFor(g, me.ID)
		in := aiseat.Input{View: protocol.ViewOfGameFor(g, me.ID.String()), Seat: me.ID, Moves: moves}
		d, err := pol.Decide(context.Background(), in)
		if err != nil {
			t.Fatalf("Decide: %v", err)
		}
		if d.Index < 0 || d.Index >= len(moves) {
			t.Fatalf("decision %d out of range of %d moves", d.Index, len(moves))
		}
		mv := moves[d.Index]
		if mv.Type == legal.TypePassPriority {
			t.Fatalf("the bot passed with the exiled Bears castable; moves %v", attackLabels(moves))
		}
		cast = mv.Type == legal.TypeCastSpell && mv.Source == bears
		if err := actions.Dispatch(g, actions.Action{
			Type: actions.Type(mv.Type), Player: mv.Player, Caller: me.ID, Params: mv.Params,
		}); err != nil {
			t.Fatalf("the engine refused the enumerated move %q: %v", mv.Label, err)
		}
	}
	if !cast {
		t.Fatal("the bot never cast the exiled Bears")
	}
	for i := 0; i < 16 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !g.Battlefield.Contains(bears) {
		t.Error("the cast Bears did not resolve onto the battlefield")
	}
	for _, p := range g.Seats {
		if want, ok := lives[p.ID]; ok && p.Life != want {
			t.Errorf("the card was cast, so no damage: opponent life %d → %d", want, p.Life)
		}
	}
}
