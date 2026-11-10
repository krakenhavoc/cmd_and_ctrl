package aiseat_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// #5401: the real Fabrication Foundry. The bot exiles artifacts whose
// total mana value is the X it announces, and returns an artifact card
// of mana value X or less from its graveyard: every move it is offered
// already pairs the payment with the targets it admits, so the
// activation it makes is one the engine accepts. The row declares a
// reanimation, so the bot prices the 8/8 it gets back against the 0/1
// it exiles, and takes the trade.
func TestHeuristicActivatesFabricationFoundry(t *testing.T) {
	g := newSettledTable(t, 5401)
	advanceToStep(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	foundry, relic, golem := uuid.New(), uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		for _, s := range g.Seats {
			s.Hand.Cards = nil
			s.Command.Cards = nil
		}
		known := map[uuid.UUID]bool{me.ID: true}
		g.Battlefield.PushTop(game.Card{InstanceID: foundry, Name: "Fabrication Foundry",
			OracleID: "0bb20c36-689f-4a7c-80b5-a12bda821847", TypeLine: "Artifact", ManaCost: "{1}{W}",
			Owner: me.ID, Controller: me.ID, KnownBy: known})
		g.Battlefield.PushTop(game.Card{InstanceID: relic, Name: "Rusted Construct", TypeLine: "Artifact Creature — Construct", ManaCost: "{3}",
			Power: 0, Toughness: 1,
			Owner: me.ID, Controller: me.ID, KnownBy: known})
		for i := 0; i < 3; i++ {
			g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains",
				Owner: me.ID, Controller: me.ID, KnownBy: known})
		}
		me.Graveyard.PushTop(game.Card{InstanceID: golem, Name: "Relic Golem", TypeLine: "Artifact Creature — Golem",
			ManaCost: "{3}", Power: 8, Toughness: 8, Owner: me.ID, Controller: me.ID})
	})
	g.BumpLayerVersionForTest()

	made := playHeliodTurn(t, g, me, heuristic.New())
	activated := false
	for _, m := range made {
		if m.move.Kind == legal.KindActivate && m.move.Source == foundry {
			activated = true
		}
	}
	if !activated {
		t.Fatal("the bot never activated Fabrication Foundry")
	}
	onField := false
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == golem && c.Controller == me.ID {
			onField = true
		}
	}
	if !onField {
		t.Error("the Golem did not return to the battlefield")
	}
	if !g.Exile.Contains(relic) {
		t.Error("the bot did not exile the three-drop to pay for X = 3")
	}
}
