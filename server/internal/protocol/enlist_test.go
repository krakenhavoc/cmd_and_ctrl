package protocol

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// enlist_test.go — #5658, the wire half of enlist (CR 702.154): the
// digest lists the creatures an attacker may enlist, read off the
// enumerator's twin moves, and EventEnlist is one log line.

func TestDigestEnlistOnAttackFromTheTwinMoves(t *testing.T) {
	enlister, plain, def, a, b := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	move := func(src uuid.UUID, enlist string) legal.Move {
		p, err := json.Marshal(map[string]any{"attacker": src.String(), "target": def.String(), "enlist": enlist})
		if err != nil {
			t.Fatal(err)
		}
		return legal.Move{Type: legal.TypeDeclareAttacker, Kind: legal.KindAttack, Source: src, Params: p}
	}
	d := digestLegalMoves([]legal.Move{
		move(enlister, ""), move(enlister, a.String()), move(enlister, b.String()), move(plain, ""),
	})
	if d == nil {
		t.Fatal("no digest")
	}
	e := d.Sources[enlister.String()]
	if len(e.EnlistOnAttack) != 2 || e.EnlistOnAttack[0] != a.String() || e.EnlistOnAttack[1] != b.String() {
		t.Errorf("enlist_on_attack = %v, want the two candidates", e.EnlistOnAttack)
	}
	if len(e.AttackTargets) != 1 {
		t.Errorf("attack_targets = %v: the twins name the same target once", e.AttackTargets)
	}
	if got := d.Sources[plain.String()].EnlistOnAttack; len(got) != 0 {
		t.Errorf("a creature with no enlist move: enlist_on_attack = %v", got)
	}
}

func TestEnlistLogLine(t *testing.T) {
	attacker, enlisted := uuid.New(), uuid.New()
	turn, step := 1, ""
	var sacrificed uuid.UUID
	e, ok := projectEvent(game.Event{Kind: game.EventEnlist, Actor: uuid.New(), CardID: attacker, Source: attacker, Target: enlisted},
		func(uuid.UUID) int { return -1 }, &turn, &step, &sacrificed)
	if !ok || e.Kind != LogEnlist || e.CardID != attacker.String() || e.Target != enlisted.String() {
		t.Fatalf("projected %+v, %v", e, ok)
	}
	if got := renderLogText(e, "Guardian of New Benalia", "Grizzly Bears"); got != "Guardian of New Benalia enlisted Grizzly Bears" {
		t.Errorf("line = %q", got)
	}
}
