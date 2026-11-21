package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// read_the_runes_test.go — the X repetitions, and the fact that each
// one is its own choice. The interesting assertions are that the two
// repetitions can be answered DIFFERENTLY and that the second question
// is built from the board the first one left behind.

const readTheRunesOracle = "74316be1-041f-4a2e-a296-0b496f90ca25"

// castReadTheRunes casts the instant for the given X.
func CastReadTheRunes(t *testing.T, g *game.Game, caster uuid.UUID, x int) {
	t.Helper()
	p := g.PlayerByIDForEffect(caster)
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Read the Runes", TypeLine: "Instant",
		OracleID: readTheRunesOracle, Owner: caster, Controller: caster,
	})
	AdvanceToMain(t, g)
	if err := g.CastSpell(caster, id, game.CastSpellParams{XValue: x}); err != nil {
		t.Fatalf("CastSpell Read the Runes: %v", err)
	}
}
