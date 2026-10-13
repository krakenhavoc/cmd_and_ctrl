package aiseat_test

import (
	"context"
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

// free_hand_cast_bot_test.go — #3270 through the real card, the real
// enumerator and the real heuristic. "You may cast a spell from your hand
// without paying its mana cost" used to grant a later cast the bot never
// took (the pick was priced as a discard). Now the bot picks the card and
// then casts it, with its target, before it passes.
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
	bears := push(game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Colors: []string{"G"}})

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

	pol := heuristic.New()
	taken := driveChoices(t, g, pol, me.ID, 4)
	if len(taken) != 1 || taken[0] == "" {
		t.Fatalf("answered the pick with %v", taken)
	}
	if !me.Hand.Contains(bears) {
		t.Fatal("the pick cast nothing by itself; the cast is the next decision")
	}

	// Next decision: cast the Bears (free), not pass.
	moves := legal.EnumerateFor(g, me.ID)
	in := aiseat.Input{View: protocol.ViewOfGameFor(g, me.ID.String()), Seat: me.ID, Moves: moves}
	d, err := pol.Decide(context.Background(), in)
	if err != nil {
		t.Fatalf("Decide: %v", err)
	}
	if d.Index < 0 || d.Index >= len(moves) {
		t.Fatalf("decision %d out of range of %d moves", d.Index, len(moves))
	}
	mv := moves[d.Index]
	if mv.Type != legal.TypeCastSpell || mv.Source != bears {
		t.Fatalf("the bot chose %q; with the free Bears on offer it should cast it; moves %v", mv.Label, attackLabels(moves))
	}
	if err := actions.Dispatch(g, actions.Action{
		Type: actions.Type(mv.Type), Player: mv.Player, Caller: me.ID, Params: mv.Params,
	}); err != nil {
		t.Fatalf("the engine refused the enumerated free cast %q: %v", mv.Label, err)
	}
	if me.Hand.Contains(bears) {
		t.Fatal("the Bears is still in hand after the cast")
	}
	for i := 0; i < 16 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !g.Battlefield.Contains(bears) {
		t.Error("the free Bears did not resolve onto the battlefield")
	}
}
