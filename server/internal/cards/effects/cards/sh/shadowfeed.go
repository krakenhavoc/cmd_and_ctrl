package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shadowfeed — {B} Instant (Shards of Alara, S82 OLD b516):
//
//	"Exile target card from a graveyard. You gain 3 life."
//
// Beckon Apparition's exile over any graveyard, then the life. If the target
// is illegal as it resolves the spell does nothing (CR 608.2b), so no life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7681ea90-be9d-4a43-845e-ba051b438429",
		Name:         "Shadowfeed",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target card from a graveyard"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.ExileTargetCards(ctx.Game, ctx.Item) }),
			effects.GainLife{Amount: 3},
		),
	})
}
