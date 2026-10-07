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
