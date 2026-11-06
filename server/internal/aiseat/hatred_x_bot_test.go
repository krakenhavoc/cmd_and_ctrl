package aiseat_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// hatred_x_bot_test.go — #4976, whole-engine: the real catalog Hatred
// ("As an additional cost to cast this spell, pay X life. Target creature
// gets +X/+0 until end of turn."), the real enumerator, view and policy.
//
// Hatred's X is announced once (CR 601.2b) and is both the life paid
// (CR 601.2f, 119.4) and the +X/+0. The enumerator offers the largest X
// the life allows, and the bot sends the move as offered, so the X the
// policy prices the pump at must be the X the engine charges and pumps
// by. The bot at 30 life attacks with a 2/2 into an opponent on 22 whose
// only creature is tapped: Hatred at X=20 (the enumerator's MaxX) makes
// the swing 22, the kill, for twenty life that leaves it on 10.

const botHatredOracle = "40601061-1b25-4db8-8b99-33324ce945cf"

func TestHeuristicPaysHatredsXInLifeForTheKill(t *testing.T) {
	g, me, them, _, attacker := heliodCombatTable(t, 4976, 0)
	hatred := game.Card{InstanceID: uuid.New(), Name: "Hatred", OracleID: botHatredOracle,
		TypeLine: "Instant", ManaCost: "{3}{B}{B}", Owner: me.ID, Controller: me.ID,
		KnownBy: map[uuid.UUID]bool{me.ID: true}}
	g.WithWriteLock(func() {
		me.Life, them.Life = 30, 22
		me.Hand.PushTop(hatred)
		for i := 0; i < 5; i++ {
			s := game.Card{InstanceID: uuid.New(), Name: "Swamp", TypeLine: "Basic Land — Swamp",
				Owner: me.ID, Controller: me.ID}
			s.AddKnowersAll(seatIDs(g))
			g.Battlefield.PushTop(s)
		}
	})
	g.BumpLayerVersionForTest()

	inHand := func() bool {
		for _, c := range me.Hand.Cards {
			if c.InstanceID == hatred.InstanceID {
				return true
			}
		}
		return false
	}
	// Stop once Hatred has left the hand and resolved, before combat
	// damage ends the game; or at the end of the turn when it is kept.
	turn := g.Turn.Seq
	cfg := heuristic.DefaultConfig()
	cfg.PriceXPumps = true // ships off (#4976's arena); the price is what this test holds
	made := playHeliodUntil(t, g, me, heuristic.NewWithConfig(cfg), func() bool {
		return (!inHand() && g.Stack.Size() == 0) || g.Turn.Seq != turn || g.Turn.Step == game.StepEnd
	})
	cast := -1
	for i, m := range made {
		if m.move.Kind == legal.KindCast && m.move.Source == hatred.InstanceID {
			cast = i
		}
	}
	if cast < 0 {
		t.Fatalf("the bot did not cast Hatred for the kill; it made %v", labelsOf(made))
	}
	var p struct {
		XValue int `json:"x_value"`
	}
	if err := json.Unmarshal(made[cast].move.Params, &p); err != nil {
		t.Fatal(err)
	}
	if p.XValue != 20 {
		t.Errorf("Hatred was cast at X=%d, want the enumerator's 20", p.XValue)
	}
	if me.Life != 30-p.XValue {
		t.Errorf("the bot is on %d life after Hatred at X=%d from 30", me.Life, p.XValue)
	}
	power := -1
	for _, c := range protocol.ViewOfGameFor(g, me.ID.String()).Battlefield.Cards {
		if c.InstanceID == attacker.String() {
			power = c.Power
		}
	}
	if power != 2+p.XValue {
		t.Errorf("the attacker's power is %d after Hatred at X=%d, want %d", power, p.XValue, 2+p.XValue)
	}
}
