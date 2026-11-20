package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rimrock Knight // Boulder Rush — Creature — Dwarf Knight {1}{R}, 3/1 // Instant — Adventure {R} (Throne of Eldraine, #5516):
//
//	Rimrock Knight — "This creature can't block."
//	Boulder Rush — "Target creature gets +2/+0 until end of turn."
//
// Both faces register: the creature under the oracle ID and the Adventure
// under "<oracle_id>#1". The Adventure lifecycle (exile on resolution, cast
// the creature later, CR 715) is the engine's.
//
// No simplification.
const rimrockKnightOracleID = "24db950b-e4fe-42d0-977b-4957eb8229d9"

func init() {
	Register(Spec{
		OracleID:     rimrockKnightOracleID,
		Name:         "Rimrock Knight",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBlock)},
	})
	Register(Spec{
		OracleID:     rimrockKnightOracleID + "#1",
		Name:         "Boulder Rush",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 0)),
		Targets:      TargetCreature("target creature"),
		OnResolve:    boostTheTargetUntilEOT(2, 0, "Boulder Rush — +2/+0 until end of turn"),
	})
}
