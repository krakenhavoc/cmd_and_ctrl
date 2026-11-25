package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Farm // Market — split card with aftermath (Commander 2019):
//
//	Farm {2}{W}, Instant: "Destroy target attacking or blocking creature."
//	Market {2}{U}, Sorcery: "Aftermath (Cast this spell only from your
//	 graveyard. Then exile it.) Draw two cards, then discard two cards."
//
// Each half is an ADR 0034 face: Farm under the bare oracle ID, Market under
// "#1"; aftermath is the engine's (CR 702.127a), as for Road // Ruin. Farm is
// Divine Verdict's destroy; Market is Faithless Looting's loot, the draws
// landing before the discard prompt.
//
// No simplification.
func init() {
	const oracleID = "2ab857cc-64b9-4115-8fa5-6dfae9edbdb5"
	effects.Register(effects.Spec{
		OracleID:     oracleID,
		Name:         "Farm",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Market",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2, Discards: 2},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.LootOne(ctx.Game, item, 2)
		},
	})
}
