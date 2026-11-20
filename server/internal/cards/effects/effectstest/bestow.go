package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bestowCard puts a bestow card in the active seat's hand and walks to
// a main phase.
func BestowCard(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return id
}

// castBestowed casts the card bestowed onto host.
func CastBestowed(t *testing.T, g *game.Game, id, host uuid.UUID) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		AlternativeCost: game.BestowKey,
		Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: host}},
	}); err != nil {
		t.Fatalf("cast bestowed: %v", err)
	}
}
