package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "dab64d0f-1246-4a4d-9a79-6db2ca0c8882",
		Name:         "Desynchronization",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BounceAllMatching{Match: And(Nonland(), Not(b14Historic()))}.Apply(ctx)
		},
	})
}
