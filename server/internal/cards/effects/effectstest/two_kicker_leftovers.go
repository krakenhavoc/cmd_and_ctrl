package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func LftSeed(g *game.Game, controller uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Power: power, Toughness: toughness,
		Owner: controller, Controller: controller,
	})
}

// lftCastKicked is castXSpell with the optional costs claimed.
func LftCastKicked(t *testing.T, g *game.Game, name, typeLine, oracle, manaCost string, x int, targets []game.TargetRef, optional []int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Targets: targets, XValue: x, OptionalCosts: optional}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// lftCastSizedCreature casts a creature whose printed P/T is a square
// `size`, so a power read off it means something.
func LftCastSizedCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, size int, optional []int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: size, Toughness: size, Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{OptionalCosts: optional}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}
