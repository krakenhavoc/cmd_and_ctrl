package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ardenvale Tactician // Dizzying Swoop — Creature — Human Knight {1}{W}{W}, 2/3 // Instant — Adventure {1}{W} (Throne of Eldraine, #5516):
//
//	Ardenvale Tactician — "Flying"
//	Dizzying Swoop — "Tap up to two target creatures."
//
// Both faces register: the creature under the oracle ID and the Adventure
// under "<oracle_id>#1". The Adventure lifecycle (exile on resolution, cast
// the creature later, CR 715) is the engine's.
//
// No simplification.
const ardenvaleTacticianOracleID = "ea27c8bf-33c5-443f-b8b2-51235efc2491"

func init() {
	Register(Spec{
		OracleID:        ardenvaleTacticianOracleID,
		Name:            "Ardenvale Tactician",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
	Register(Spec{
		OracleID:     ardenvaleTacticianOracleID + "#1",
		Name:         "Dizzying Swoop",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return tapEachLegalTarget(ctx.Game, item)
		},
	})
}
