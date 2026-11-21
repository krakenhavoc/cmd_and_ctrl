package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sacrifice_cost_test.go — the sacrifice half of CR 601.2f additional
// costs: Village Rites, Altar's Reap, Deadly Dispute. The discard half
// lives in additional_cost_test.go.
//
// The interesting assertions here are about ORDER, not about drawing
// cards. An additional cost is paid while the spell is being cast, so
// the sacrificed creature is already dead before the spell is on the
// stack. Everything downstream of that — aristocrats triggers landing
// above the spell, the creature staying dead through a counterspell,
// the cast being illegal with nothing to eat — follows from it, and
// each of those is a way an implementation that sacrificed on
// resolution instead would look correct until someone played the deck.

const (
	VillageRitesOracle  = "365548fb-5acc-4a8a-b20b-26d28b7d029f"
	AltarsReapOracle    = "6a125750-2b8c-4f9d-8173-ac8d14c91ddb"
	DeadlyDisputeOracle = "457af74a-02b3-4659-846d-63e482667f34"
)

// castWithSacrifice casts a catalog spell from the active player's
// hand, paying an additional sacrifice cost with the named permanent.
// Returns the spell's instance ID; fatals if the cast is rejected.
func CastWithSacrifice(t *testing.T, g *game.Game, name, typeLine, oracleID string, victim uuid.UUID) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		Owner:      active.ID,
		Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{victim},
	}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// tryCastWithSacrifice is castWithSacrifice without the fatal, for the
// rejection cases.
func TryCastWithSacrifice(g *game.Game, name, typeLine, oracleID string, chosen []uuid.UUID) (uuid.UUID, error) {
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		Owner:      active.ID,
		Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			return id, err
		}
	}
	return id, g.CastSpell(active.ID, id, game.CastSpellParams{SacrificeIDs: chosen})
}
