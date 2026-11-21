package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Giant Killer // Chop Down — Creature — Human Peasant {W}, 1/2 // Instant — Adventure {2}{W} (Throne of Eldraine, #5516):
//
//	Giant Killer — "{1}{W}, {T}: Tap target creature."
//	Chop Down — "Destroy target creature with power 4 or greater."
//
// Both faces register: the creature under the oracle ID and the Adventure
// under "<oracle_id>#1". The Adventure lifecycle (exile on resolution, cast
// the creature later, CR 715) is the engine's.
//
// Chop Down is Bring Down's text without assist.
//
// No simplification.
const giantKillerOracleID = "91e71279-c9d2-4873-916a-59da03a65741"

func init() {
	effects.Register(effects.Spec{
		OracleID:     giantKillerOracleID,
		Name:         "Giant Killer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     giantKillerOracleID + "#1",
		Name:         "Chop Down",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature with power 4 or greater", effects.PowerGE(4)),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
