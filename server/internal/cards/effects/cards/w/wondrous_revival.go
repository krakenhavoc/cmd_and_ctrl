package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wondrous Revival — Sorcery {4}{W}{W} (Marvel Super Heroes Commander,
// #4455):
//
//	"Return up to three target Hero creature cards from your graveyard to
//	 the battlefield."
//
// Reveillark's return (b24ReturnGraveyardTargetsToBattlefield): the
// still-legal targets enter together. Hero is a creature subtype here.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "19cab267-3c51-4322-b1af-22941bbbdadf",
		Name:         "Wondrous Revival",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("up to three target Hero creature cards from your graveyard",
			effects.YouOwn(), effects.Creature(), effects.Subtype("Hero")).WithCount(0, 3),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B24ReturnGraveyardTargetsToBattlefield(ctx, 3)
		},
	})
}
