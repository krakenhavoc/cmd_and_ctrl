package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dogpile — Instant {3}{R} (Ravnica: City of Guilds, #5600):
//
//	"Dogpile deals damage to any target equal to the number of attacking
//	 creatures you control."
//
// The count is Drana's "attacking creatures you control", read as the
// spell resolves; with none attacking it deals no damage (CR 120.8). A
// target gone by then makes it do nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a4364218-f746-42ff-8495-bf8ca69c3b62",
		Name:         "Dogpile",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return dogpileDamage(ctx.Game, item)
		},
	})
}

// dogpileDamage is the spell's damage, counted once as it resolves.
var dogpileDamage = sourceDealsComputedDamageToEachLegalTarget(func(ctx *Context) int {
	return len(b13AttackingCreaturesYouControl(ctx.Game, ctx.Controller()))
})
