package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seismic Strike — Instant {2}{R} (Magic 2010, #5585):
//
//	"Seismic Strike deals damage to target creature equal to the number
//	 of Mountains you control."
//
// Rumbling Rockslide's shape counting Mountains: the count is taken as the
// spell resolves, so a Mountain that left in response is not counted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3513d862-ef0d-4949-b33a-3fa9437cc444",
		Name:         "Seismic Strike",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
			mountains := effects.B08LandsWithSubtypeControlled(ctx.Game, ctx.Controller(), ctx.TextSubtype("Mountain"))
			return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: mountains}.Apply(ctx)
		})),
	})
}
