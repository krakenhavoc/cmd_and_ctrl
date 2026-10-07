package protocol

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// exert_test.go — ADR 0130 §5: the wire half of exert. The digest says
// which creatures may be exerted as they attack, read off the
// enumerator's twin move, and EventExert is one log line.

func exertDigestMove(t *testing.T, src, target uuid.UUID, exert bool) legal.Move {
	t.Helper()
	p, err := json.Marshal(map[string]any{"attacker": src.String(), "target": target.String(), "exert": exert})
	if err != nil {
		t.Fatal(err)
	}
	return legal.Move{Type: legal.TypeDeclareAttacker, Kind: legal.KindAttack, Source: src, Params: p}
}

func TestDigestExertOnAttackFromTheTwinMove(t *testing.T) {
	exerter, plain, def := uuid.New(), uuid.New(), uuid.New()
	d := digestLegalMoves([]legal.Move{
		exertDigestMove(t, exerter, def, false),
		exertDigestMove(t, exerter, def, true),
		exertDigestMove(t, plain, def, false),
	})
	if d == nil {
		t.Fatal("no digest")
	}
	e := d.Sources[exerter.String()]
	if !e.ExertOnAttack {
		t.Error("a creature with an exert move: exert_on_attack")
	}
	if len(e.AttackTargets) != 1 || e.AttackTargets[0] != def.String() {
		t.Errorf("attack_targets = %v: the twin names the same target once", e.AttackTargets)
	}
	if d.Sources[plain.String()].ExertOnAttack {
		t.Error("a creature with no exert move: no exert_on_attack")
	}
}

func TestExertLogLine(t *testing.T) {
	card := uuid.New()
	turn, step := 1, ""
	var sacrificed uuid.UUID
	e, ok := projectEvent(game.Event{Kind: game.EventExert, Actor: uuid.New(), CardID: card, Source: card, Target: uuid.New()},
		func(uuid.UUID) int { return 0 }, &turn, &step, &sacrificed)
	if !ok || e.Kind != LogExert || e.CardID != card.String() {
		t.Fatalf("projected %+v, %v", e, ok)
	}
	e.actorName = "Alice"
	if got := renderLogText(e, "Oketra's Avenger", ""); got != "Alice exerted Oketra's Avenger" {
		t.Errorf("line = %q", got)
	}
}

// TestExertRowsOnTheWire is ADR 0130's amendment of 2026-10-07 on the
// wire: an exert card's rows say which of exert's triggers they are, and
// carry the purpose the card declares, for the bot to price.
func TestExertRowsOnTheWire(t *testing.T) {
	owner := uuid.New()
	cases := []struct {
		name, typeLine, oracle string
		exert, purpose         string
	}{
		{"Oketra's Avenger", "Creature — Human Warrior", "8f064160-3afe-408a-85b4-b335eae8571c", "linked", `{"prevent_combat_damage_to_self":true}`},
		{"Combat Celebrant", "Creature — Human Warrior", "5e15ff93-99a0-4000-918e-4bd2c257188d", "linked", `{"extra_combat":1}`},
		{"Glorybringer", "Creature — Dragon", "b75c3902-633e-4d24-acde-d7a9cc8f466e", "linked", `{"damage_to_creature":4}`},
		{"Resolute Survivors", "Creature — Human Warrior", "3d699db7-cc52-4ee3-947b-0ba291bf037a", "payoff", `{"damage_each_opponent":1,"life_gain":1}`},
	}
	for _, c := range cases {
		rows := viewOfAbilityRows(rowsFixtureCard(c.name, c.typeLine, c.oracle, owner))
		var got []AbilityRowView
		for _, r := range rows {
			if r.Exert != "" {
				got = append(got, r)
			}
		}
		if len(got) != 1 {
			t.Errorf("%s: %d exert rows in %+v, want 1", c.name, len(got), rows)
			continue
		}
		if got[0].Exert != c.exert {
			t.Errorf("%s: exert = %q, want %q", c.name, got[0].Exert, c.exert)
		}
		b, err := json.Marshal(got[0].Purpose)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != c.purpose {
			t.Errorf("%s: purpose = %s, want %s", c.name, b, c.purpose)
		}
	}
}

// TestPumpOnTheWire: a declared pump projects as {power, toughness,
// keywords}, and its keyword list is a copy.
func TestPumpOnTheWire(t *testing.T) {
	kw := []string{"trample"}
	v := viewOfPurpose(game.Purpose{Pump: &game.Pump{Power: 2, Keywords: kw}})
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"pump":{"power":2,"keywords":["trample"]}}` {
		t.Errorf("pump = %s", b)
	}
	kw[0] = "flying"
	if v.Pump.Keywords[0] != "trample" {
		t.Error("the view shares the catalog's keyword slice")
	}
}
