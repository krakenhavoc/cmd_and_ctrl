package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Synchronized Eviction — Instant {4}{U} (Jumpstart 2022, #5948):
//
//	"This spell costs {2} less to cast if you control at least two
//	 creatures that share a creature type.
//	 Put target nonland permanent into its owner's library second from
//	 the top."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad64b05a-4702-47ef-a81e-1b481e57ab55",
		Name:         "Synchronized Eviction",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(2, "This spell costs {2} less to cast if you control at least two creatures that share a creature type.",
				effects.YouControlTwoCreaturesSharingAType()),
		},
		OnResolve: effects.PutTargetPermanentSecondFromTheTop,
	})
}
