package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jolt — Instant {2}{U} (Mirage, S82 OLD b800, #5633):
//
//	"You may tap or untap target artifact, creature, or land.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// Bounding Krasis's tap-or-untap choice (tap, untap or neither, asked as the
// spell resolves) over an artifact, creature or land; the draw is Burnout's
// delayed trigger at the next upkeep, scheduled whatever the choice was. A
// target gone by resolution does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "13f69c70-dc93-4983-b649-d91b22633eb9",
		Name:         "Jolt",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact, creature, or land", effects.Or(effects.Artifact(), effects.Creature(), effects.Land())),
		Purpose:      game.Purpose{Answers: game.AnswerValue, Draws: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.NemMayTapOrUntapTheTarget("Jolt")(ctx.Game, item); err != nil {
				return err
			}
			return effects.DrawACardAtTheNextUpkeep("Jolt — draw a card").Apply(ctx)
		},
	})
}
