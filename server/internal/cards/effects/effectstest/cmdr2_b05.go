package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func Cmdr2B05Card(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tt int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tt, Owner: owner, Controller: owner,
	})
}

// --- Kopala ---------------------------------------------------------

const Cmdr2B05ProbeOracle = "test-cmdr2-b05-probe"

func init() {
	// {1}: target creature. A probe ability, usable by any player.
	effects.Register(effects.Spec{
		OracleID: Cmdr2B05ProbeOracle,
		Name:     "B05 Probe",
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: Target creature does nothing.",
			Cost:    effects.ManaCost("{1}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
}
