package aiseat_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// disguise_x_bot_test.go — #5249 through the real engine: the bot turns a
// disguise {X} creature face up with an X it chose. Its rule is the
// enumerator's: the largest X it can pay.

const oracleAureliasVindicatorBot = "4f452ce7-1166-4810-afed-ea5e156a6f52"

// The bot holds six mana over a face-down Aurelia's Vindicator
// (disguise {X}{3}{W}): it turns it up for X = 2, the most it can pay,
// and the engine takes the move and charges all six.
func TestHeuristicTurnsADisguiseXCreatureUpForTheMaxAffordableX(t *testing.T) {
	g, me := tbTable(t, 5249, "{C}{C}{C}")
	av := game.NewCard("Aurelia's Vindicator", me.ID)
	av.TypeLine, av.ManaCost, av.OracleID = "Creature — Angel", "{2}{W}{W}", oracleAureliasVindicatorBot
	av.Power, av.Toughness = 4, 2
	id := tbHandCard(g, me, av)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "disguise"}); err != nil {
		t.Fatalf("cast face down: %v", err)
	}
	for i := 0; i < 16 && len(g.StackMeta) > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority %d: %v", i, err)
		}
	}
	if len(g.StackMeta) != 0 {
		t.Fatal("the face-down Vindicator did not resolve")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{C}{C}{C}{C}{C}"); err != nil {
		t.Fatal(err)
	}

	m := tbDecide(t, g, me, heuristic.DefaultConfig())
	if !tbTurnsFaceUp(t, m) || m.Source != id {
		t.Fatalf("the bot chose %q; want to turn the Vindicator face up", m.Label)
	}
	var p struct {
		XValue int `json:"x_value"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatal(err)
	}
	if p.XValue != 2 {
		t.Errorf("the bot chose X=%d; six mana pay {2}{3}{W}, so the max affordable X is 2", p.XValue)
	}
	if m.Value == nil || m.Value.Kind != legal.ValueX || m.Value.Max == nil || *m.Value.Max != 2 {
		t.Errorf("move value %+v, want an open X up to 2", m.Value)
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: me.ID, Params: m.Params}); err != nil {
		t.Fatalf("dispatch %q: %v", m.Label, err)
	}
	var faceDown bool
	var left int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				faceDown = c.FaceDown
			}
		}
		left = me.ManaPool.Total()
	})
	if faceDown || left != 0 {
		t.Errorf("after the move: face down %v, %d mana left; want face up and all six spent", faceDown, left)
	}
}
