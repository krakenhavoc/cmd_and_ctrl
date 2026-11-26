package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Restock — Sorcery {3}{G}{G} (Magic 2015, #5556):
//
//	"Return two target cards from your graveyard to your hand. Exile Restock."
//
// Retrieve's shape: both targets are checked again on resolution (CR 608.2b)
// and each still legal one returns; the spell then moves itself to exile
// instead of the graveyard (#489). A spell countered or with every target
// illegal is not exiled, because it never resolved.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8b641cd3-bb08-4f2e-9dad-24f58a9c32e1",
		Name:         "Restock",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetCardInGraveyard("two target cards from your graveyard", effects.YouOwn()).WithCount(2, 2),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.ReturnLegalGraveyardTargetsToHand),
			effects.Instruction(func(ctx *effects.Context) error {
				return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
			}),
		),
	})
}
