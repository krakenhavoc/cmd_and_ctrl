package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        ardenvaleTacticianOracleID,
		Name:            "Ardenvale Tactician",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
	effects.Register(effects.Spec{
		OracleID:     ardenvaleTacticianOracleID + "#1",
		Name:         "Dizzying Swoop",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TapEachLegalTarget(ctx.Game, item)
		},
	})
}
