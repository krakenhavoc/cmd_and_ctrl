package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// modes_test.go — S20 sub-PR 4: modal cards resolve the chosen
// option only, and the option's target clause gates the cast.

const (
	RakdosCharmOracle    = "5e62b51d-faec-4aa0-9504-cf2c282d08ea"
	IzzetCharmOracle     = "a07698f6-5ad5-49a3-9da2-f82d407f5cd7"
	AustereCommandOracle = "09cc8709-fe10-472a-b05c-e89f3523018d"
)

// castModal seeds a modal card into the active seat's hand and casts
// it with the chosen modes + targets.
func CastModal(t *testing.T, g *game.Game, name, typeLine, oracle string, modes []int, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

func PushCostedPermanentForTest(g *game.Game, owner uuid.UUID, name, typeLine, manaCost string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
	return id
}
