package legal_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// exert_test.go — ADR 0130 §6, test 11's enumerator half: an attacker
// that may be exerted as it attacks is offered each attack twice, the
// plain one and the same one with `exert: true`, and the dispatcher
// accepts both (#544).

const (
	oracleOketrasAvenger = "8f064160-3afe-408a-85b4-b335eae8571c"
	oracleCombatCeleb    = "5e15ff93-99a0-4000-918e-4bd2c257188d"
)

type exertParams struct {
	Attacker string `json:"attacker"`
	Target   string `json:"target"`
	Exert    bool   `json:"exert"`
}

// attackMovesOf splits the creature's attack moves into plain and
// exerting, by target.
func attackMovesOf(t *testing.T, moves []legal.Move, id uuid.UUID) (plain, exert map[string]legal.Move) {
	t.Helper()
	plain, exert = map[string]legal.Move{}, map[string]legal.Move{}
	for _, m := range moves {
		if m.Kind != legal.KindAttack || m.Source != id {
			continue
		}
		var p exertParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		if p.Exert {
			exert[p.Target] = m
		} else {
			plain[p.Target] = m
		}
	}
	return plain, exert
}

func TestExertCreatureIsOfferedTwinAttackMoves(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	avenger := battlefieldCard(g, active, game.Card{
		Name: "Oketra's Avenger", TypeLine: "Creature — Human Warrior", ManaCost: "{1}{W}",
		Power: 3, Toughness: 1, OracleID: oracleOketrasAvenger,
	})
	bear := battlefieldCard(g, active, creature("Bear", "{G}", 2, 2))
	advanceTo(t, g, game.StepDeclareAttackers)

	moves := legal.EnumerateFor(g, active.ID)
	plain, exert := attackMovesOf(t, moves, avenger)
	if len(plain) == 0 || len(plain) != len(exert) {
		t.Fatalf("%d plain and %d exerting attacks, want one of each per target", len(plain), len(exert))
	}
	for target, m := range exert {
		if _, ok := plain[target]; !ok {
			t.Errorf("an exert move at %s with no plain twin", target)
		}
		if !strings.Contains(m.Label, "and exert it (it won't untap during your next untap step)") {
			t.Errorf("exert move label %q", m.Label)
		}
		if m.AlwaysLegal {
			t.Error("an exert move is never the owed answer (CR 508.1d)")
		}
	}
	if _, ex := attackMovesOf(t, moves, bear); len(ex) != 0 {
		t.Error("a creature without the ability is offered no exert")
	}
	dispatchAll(t, g, active.ID, moves)

	// Exerting one: the twin goes with the attack, and the creature is
	// exerted at the lock-in.
	for _, m := range exert {
		var p exertParams
		_ = json.Unmarshal(m.Params, &p)
		target := uuid.MustParse(p.Target)
		if err := g.DeclareAttackerDeclWith(game.AttackDeclaration{Attacker: avenger, Target: target, Exert: true}, game.DeclareAttackersParams{}); err != nil {
			t.Fatal(err)
		}
		break
	}
	if _, ex := attackMovesOf(t, legal.EnumerateFor(g, active.ID), avenger); len(ex) != 0 {
		t.Error("a declared attacker is offered no more attacks")
	}
}

// TestCombatCelebrantIsNotOfferedASecondExert — "If this creature
// hasn't been exerted this turn": once it has, only the plain attack.
func TestCombatCelebrantIsNotOfferedASecondExert(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	def := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	celebrant := battlefieldCard(g, active, game.Card{
		Name: "Combat Celebrant", TypeLine: "Creature — Human Warrior", ManaCost: "{2}{R}",
		Power: 4, Toughness: 1, OracleID: oracleCombatCeleb,
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	if _, ex := attackMovesOf(t, legal.EnumerateFor(g, active.ID), celebrant); len(ex) == 0 {
		t.Fatal("Combat Celebrant is offered an exert before it has been exerted")
	}
	if err := g.DeclareAttackerDeclWith(game.AttackDeclaration{Attacker: celebrant, Target: def.ID, Exert: true}, game.DeclareAttackersParams{}); err != nil {
		t.Fatal(err)
	}
	// Lock in, resolve the trigger, untap Celebrant by hand and walk to
	// the added combat.
	for g.Turn.Step == game.StepDeclareAttackers {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	advanceTo(t, g, game.StepEndCombat)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	if g.Turn.Step != game.StepBeginCombat {
		t.Fatalf("after the first combat: %s, want the added combat", g.Turn.Step)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == celebrant {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	plain, ex := attackMovesOf(t, legal.EnumerateFor(g, active.ID), celebrant)
	if len(plain) == 0 {
		t.Fatal("Celebrant may still attack")
	}
	if len(ex) != 0 {
		t.Error("Celebrant has been exerted this turn; no exert is offered")
	}
}
