package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ADR 0187 §4 (#3569): CardView.AssignsNoCombatDamage is stamped from the
// live scoped effect, ships as assigns_no_combat_damage, is public, and
// round-trips.
func TestViewCarriesAssignsNoCombatDamage(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	rats := game.NewCard("Carrion Rats", me.ID)
	rats.TypeLine = "Creature — Rat"
	rats.Power, rats.Toughness = 2, 1
	bear := game.NewCard("Grizzly Bears", me.ID)
	bear.TypeLine = "Creature — Bear"
	bear.Power, bear.Toughness = 2, 2
	g.Battlefield.PushTop(rats)
	g.Battlefield.PushTop(bear)
	var ok bool
	g.WithWriteLock(func() { ok = g.AssignsNoCombatDamageThisTurnForEffect(uuid.Nil, rats.InstanceID, "Carrion Rats") })
	if !ok {
		t.Fatal("setup: no record written")
	}

	for _, viewer := range []string{me.ID.String(), g.Seats[1].ID.String()} {
		v := ViewOfGameFor(g, viewer)
		var gotRats, gotBear *CardView
		for i := range v.Battlefield.Cards {
			switch v.Battlefield.Cards[i].InstanceID {
			case rats.InstanceID.String():
				gotRats = &v.Battlefield.Cards[i]
			case bear.InstanceID.String():
				gotBear = &v.Battlefield.Cards[i]
			}
		}
		if gotRats == nil || gotBear == nil {
			t.Fatalf("viewer %s: cards missing from the battlefield view", viewer)
		}
		if !gotRats.AssignsNoCombatDamage {
			t.Errorf("viewer %s: the marked creature does not carry assigns_no_combat_damage", viewer)
		}
		if gotBear.AssignsNoCombatDamage {
			t.Errorf("viewer %s: the unmarked creature carries assigns_no_combat_damage", viewer)
		}

		raw, err := json.Marshal(gotRats)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(raw), `"assigns_no_combat_damage":true`) {
			t.Errorf("wire form lacks the field: %s", raw)
		}
		var back CardView
		if err := json.Unmarshal(raw, &back); err != nil {
			t.Fatal(err)
		}
		if !back.AssignsNoCombatDamage {
			t.Error("the field did not round-trip")
		}
		rawBear, _ := json.Marshal(gotBear)
		if strings.Contains(string(rawBear), "assigns_no_combat_damage") {
			t.Errorf("an unmarked creature ships the field: %s", rawBear)
		}
	}
}
