package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     giantKillerOracleID,
		Name:         "Giant Killer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}, {T}: Tap target creature.",
			Cost:    Plus(ManaCost("{1}{W}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
	Register(Spec{
		OracleID:     giantKillerOracleID + "#1",
		Name:         "Chop Down",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature with power 4 or greater", PowerGE(4)),
		OnResolve:    destroyTheTargetPermanent,
	})
}
