package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Desynchronization — Instant {2}{U}{U} (EDHREC rank 1574):
//
//	"Return each nonland permanent that's not historic to its owner's
//	 hand. (Artifacts, legendaries, and Sagas are historic.)"
//
// A one-sided Evacuation for the artifact-and-legends deck. The
// sweep is BounceAllMatching over "nonland and not historic", with
// historic read as CR 205.4h has it — an artifact, a legendary
// permanent, or a Saga, off the effective characteristic so a
// type-changed permanent is judged as it is. Tokens bounced this way
// cease to exist; everyone's permanents are swept, the caster's
// included, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID: "dab64d0f-1246-4a4d-9a79-6db2ca0c8882",
		Name:     "Desynchronization",
		// ADR 0126 §6: a historic permanent stays.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepBounce, Partial: true}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BounceAllMatching{Match: effects.And(effects.Nonland(), effects.Not(effects.B14Historic()))}.Apply(ctx)
		},
	})
}
