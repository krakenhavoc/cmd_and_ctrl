package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Undying Evil — Instant {B}:
//
//	"Target creature gains undying until end of turn."
//
// The grant is a layer-6 record pinned to the creature, read off its
// last-known ability list if it dies this turn (game/undying_persist.go,
// #2075). A creature that already has undying gets a second instance:
// both trigger, and the first to resolve returns it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ceaeb1b-8b25-453d-ac09-e466cfdd9c77",
		Name:         "Undying Evil",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return grantEachLegalTargetUntilEOT(ctx, game.KeywordUndying, "Undying Evil — undying until end of turn")
		},
	})
}
