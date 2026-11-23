package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// castInPlace casts an instant from caster's hand WITHOUT advancing
// the turn, so the current step (and any combat declarations) survive
// into resolution.
func CastInPlace(t *testing.T, g *game.Game, caster uuid.UUID, name, oracle string) uuid.UUID {
	t.Helper()
	var p *game.Player
	for _, s := range g.Seats {
		if s != nil && s.ID == caster {
			p = s
		}
	}
	if p == nil {
		t.Fatalf("no such seat %s", caster)
	}
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Instant",
		OracleID: oracle, Owner: caster, Controller: caster,
	})
	if err := g.CastSpell(caster, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}
