package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// resolution_keyword_cast_bot_test.go — #4243 through the real card, the
// real enumerator and the real heuristic. Cascade's cast is a
// resolution_cast prompt, answered inside the cascade trigger's
// resolution: the enumerator offers the casts of the hit and the
// always-legal decline, the heuristic must answer (a seat that owes a
// choice and is offered a move it never takes wedges the table, #544),
// and the engine accepts what it chose.
//
// Not behind AISEAT_GAME_TESTS: one resolution, a few decisions.

const botShardlessAgentOracle = "2afbaa9a-c171-4a8b-90f3-5250d8498356"

func TestHeuristicSeatAnswersACascadeResolutionCast(t *testing.T) {
	g := newSettledTable(t, 4243)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToStep(t, g, game.StepPrecombatMain)
	me.Hand.Cards = nil
	me.Command.Cards = nil

	agent := uuid.New()
	hand := game.Card{InstanceID: agent, Name: "Shardless Agent", OracleID: botShardlessAgentOracle,
		TypeLine: "Artifact Creature — Human Rogue", ManaCost: "{2}{G}", Power: 2, Toughness: 2,
		Colors: []string{"G"}, Owner: me.ID, Controller: me.ID}
	hand.AddKnowersAll(seatIDs(g))
	me.Hand.PushTop(hand)
	for i := 0; i < 3; i++ {
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
			Owner: me.ID, Controller: me.ID})
	}
	bears := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: bears, Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})

	if err := g.CastSpell(me.ID, agent, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the Agent: %v", err)
	}
	for i := 0; i < 16 && !owesChoice(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !owesChoice(g, me.ID) {
		t.Fatal("cascade asked nothing")
	}
	if !g.ResolutionPaused() {
		t.Fatal("the offer is not inside the cascade trigger's resolution")
	}

	// Both answers are offered, and the decline is the always-legal one.
	var casts, declines int
	for _, mv := range legal.EnumerateFor(g, me.ID) {
		switch {
		case mv.ResolutionCast == "":
			t.Errorf("move %q does not answer the cascade offer", mv.Label)
		case mv.Kind == legal.KindCast && mv.Source == bears:
			casts++
		case mv.Kind == legal.KindChoice && mv.AlwaysLegal:
			declines++
		}
	}
	if casts == 0 || declines != 1 {
		t.Fatalf("offered %d casts of the hit and %d declines", casts, declines)
	}

	taken := driveChoices(t, g, heuristic.New(), me.ID, 4)
	if len(taken) != 1 {
		t.Fatalf("answered the cascade offer with %v", taken)
	}
	if !g.Stack.Contains(bears) {
		t.Errorf("the bot did not cast the free two-drop (took %v)", taken)
	}
	if g.ResolutionPaused() {
		t.Error("the cascade trigger is still held after the bot's answer")
	}
	for i := 0; i < 16 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !g.Battlefield.Contains(bears) || !g.Battlefield.Contains(agent) {
		t.Error("the cascaded Bears and the Agent did not both resolve")
	}
}
