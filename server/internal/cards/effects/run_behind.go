package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Run Behind — Instant {3}{U} (Secrets of Strixhaven, #4456):
//
//	"This spell costs {1} less to cast if it targets an attacking creature.
//	 Target creature's owner puts it on their choice of the top or bottom of
//	 their library."
//
// The discount is Depower's CostsLessIfItTargets over an attacking creature.
// The placement is the OWNER's choice, not the caster's: PutInLibraryInAnyOrder
// with Chooser set to the creature's owner and the top-or-bottom lane (Endless
// Detour's placement), so the prompt is addressed to that player. A creature
// that is gone or is no longer a legal target does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7578a192-82e5-43ff-8dee-225f313c484a",
		Name:         "Run Behind",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(1, "This spell costs {1} less to cast if it targets an attacking creature.",
				And(Creature(), attackingCreature())),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ownerTucksFirstTargetTopOrBottom(ctx, "Run Behind — put it on your choice of the top or bottom of your library")
		},
	})
}
