package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fang Dragon // Forktail Sweep — an adventure card (Commander Legends: Battle for
// Baldur's Gate):
//
//	Fang Dragon {5}{R}{R}, Creature — Dragon 6/3:
//	"Flying"
//
//	Forktail Sweep {1}{R}, Sorcery — Adventure:
//	"Forktail Sweep deals 1 damage to each creature you don't control. (Then exile
//	 this card. You may cast the creature later from exile.)"
//
// The adventure layout is Brazen Borrower's (CR 715.3d). The sweep is Artistic
// Process's damage to each creature the caster doesn't control, all at once
// (CR 120.3), declared as a damage sweep for the bots.
//
// No simplification.
func init() {
	const oracleID = "c075637a-e94c-4b8d-8a9e-20ee3dfa8fe7"

	effects.Register(effects.Spec{
		OracleID:        oracleID,
		Name:            "Fang Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})

	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Forktail Sweep",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.OpponentControls()), 1)
		},
	})
}
