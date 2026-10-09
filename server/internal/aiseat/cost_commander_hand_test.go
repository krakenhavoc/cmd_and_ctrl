package aiseat_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// cost_commander_hand_test.go — #2420 through the real engine. A
// commander returned to its owner's hand to PAY A COST is asked CR
// 903.9b before anything is paid (#1397), and since #2420 that prompt
// says where a "no" sends it (playable_from_zone), as the question
// about an effect's bounce does (#2390). The heuristic keeps the
// commander in hand, and casts it from there later for its printed
// cost: no CR 903.8 tax, and no command-zone cast counted.
//
// Not behind AISEAT_GAME_TESTS: one activation, one prompt, one cast.

func TestHeuristicKeepsACommanderReturnedToPayACostInHand(t *testing.T) {
	g := newSettledTable(t, 2420)
	advanceToStep(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	everyone := seatIDs(g)

	var cmd, src uuid.UUID
	g.WithWriteLock(func() {
		// Only the commander will be in hand, so the one spell the bot
		// can cast later is the one under test.
		me.Hand.Cards = nil

		c := game.NewCommander("Test Commander", me.ID)
		c.TypeLine = "Legendary Creature — Elf"
		c.ManaCost = "{1}{R}"
		c.Power, c.Toughness = 2, 2
		c.Controller = me.ID
		c.AddKnowersAll(everyone)
		g.Battlefield.PushTop(c)
		cmd = c.InstanceID

		outlet := game.NewCard("Return Outlet", me.ID)
		outlet.TypeLine = "Artifact"
		outlet.Controller = me.ID
		outlet.ActivatedAbilities = []game.ActivatedAbilityShape{{
			Label: "Return a creature you control to its owner's hand: nothing",
			Cost: game.AbilityCost{ReturnToHand: &game.ReturnToHandCost{
				Count: 1,
				Label: "a creature you control",
				Filter: &game.TargetSpec{
					Mode:  "permanent",
					Label: "a creature you control",
					Zones: []game.ZoneKind{game.ZoneBattlefield},
					CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
						return c.IsCreature()
					},
					Min: 1, Max: 1,
				},
			}},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}}
		outlet.AddKnowersAll(everyone)
		g.Battlefield.PushTop(outlet)
		src = outlet.InstanceID
	})

	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{ReturnIDs: []uuid.UUID{cmd}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	v := protocol.ViewOfGameFor(g, me.ID.String())
	if len(v.PendingChoices) != 1 {
		t.Fatalf("pending choices on the wire = %d, want the one CR 903.9b prompt", len(v.PendingChoices))
	}
	if ch := v.PendingChoices[0]; ch.Kind != string(game.PendingChoiceOptionalReplacement) || !ch.PlayableFromZone {
		t.Fatalf("prompt %q playable_from_zone=%v, want an optional_replacement saying the commander is headed for a hand", ch.Kind, ch.PlayableFromZone)
	}

	taken := driveChoices(t, g, heuristic.New(), me.ID, 4)
	if !me.Hand.Contains(cmd) || me.Command.Contains(cmd) {
		t.Fatalf("the commander is not in hand (took %v): the bot should keep it there, where it is cast without the tax", taken)
	}

	// Resolve the ability, then give the bot exactly the printed cost.
	for i := 0; i < 16 && len(g.StackMeta) > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority %d: %v", i, err)
		}
	}
	if len(g.StackMeta) != 0 {
		t.Fatal("the ability did not resolve")
	}
	if g.Turn.Step != game.StepPrecombatMain || g.Seats[g.Turn.ActiveSeat].ID != me.ID {
		t.Fatalf("the table moved on to %s; the bot must still be in its main phase", g.Turn.Step)
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{R}"); err != nil {
		t.Fatal(err)
	}

	moves := legal.EnumerateFor(g, me.ID)
	in := aiseat.Input{View: protocol.ViewOfGameFor(g, me.ID.String()), Seat: me.ID, Moves: moves}
	d, err := heuristic.New().Decide(context.Background(), in)
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	if d.Index < 0 || d.Index >= len(moves) {
		t.Fatalf("the bot made no move (index %d of %d)", d.Index, len(moves))
	}
	mv := moves[d.Index]
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(mv.Type), Player: mv.Player, Caller: me.ID, Params: mv.Params}); err != nil {
		t.Fatalf("dispatch %q: %v", mv.Label, err)
	}
	if !g.Stack.Contains(cmd) {
		t.Fatalf("the bot chose %q; want it to cast the commander from hand for its printed {1}{R}", mv.Label)
	}
	if n := me.CommanderCasts[cmd]; n != 0 {
		t.Errorf("CommanderCasts = %d: a cast from hand is not a cast from the command zone (CR 903.8)", n)
	}
	if left := len(me.ManaPool); left != 0 {
		t.Errorf("%d mana left in the pool after the cast, want the printed cost spent exactly", left)
	}
}
