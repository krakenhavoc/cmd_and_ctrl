package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cackling Flames — Instant {3}{R} (Dissension, #5598):
//
//	"Cackling Flames deals 3 damage to any target.
//	 Hellbent — Cackling Flames deals 5 damage instead if you have no cards in
//	 hand."
//
// Galvanic Blast's shape with Demonfire's SpellHellbent, read as the
// spell resolves: by then the Flames have left the hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ab48147c-0bd7-4875-8e11-0e26185453dd",
		Name:         "Cackling Flames",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := 3
			if SpellHellbent()(ctx.Game, item) {
				n = 5
			}
			return damageToFirstTarget(n)(item, ctx)
		},
	})
}
