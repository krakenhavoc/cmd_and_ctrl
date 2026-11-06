package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blooming Blast — Instant {1}{R} (Bloomburrow, #4467):
//
//	"Gift a Treasure (You may promise an opponent a gift as you cast
//	 this spell. If you do, they create a Treasure token before its
//	 other effects. It's an artifact with "{T}, Sacrifice this token:
//	 Add one mana of any color.")
//	 Blooming Blast deals 2 damage to target creature. If the gift was
//	 promised, Blooming Blast also deals 3 damage to that creature's
//	 controller."
//
// The gift keyword (gift.go, CR 702.174): the chosen opponent's Treasure
// is made first. The 2 damage and the 3 are separate steps, so the
// second waits for any prompt the first asks (CR 608.2c, #4155). "That
// creature's controller" is read as the second step runs, off the
// creature (still on the battlefield, since state-based actions wait for
// the spell to finish). A target that has left counters the spell (CR
// 608.2b), and nothing is dealt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2e5b8973-4f6e-4abe-bbbe-81fdca784997",
		Name:         "Blooming Blast",
		Completeness: CompletenessFull,
		Gift:         GiftATreasure(),
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 2)), Answers: game.AnswerRemove},
		OnResolve: ResolveSteps(
			damageTheClauseTarget(0, 2),
			Instruction(func(ctx *Context) error {
				if !ctx.GiftPromised() {
					return nil
				}
				t, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				c, found := ctx.Game.LookupCardForEffect(t.ID)
				if !found {
					return nil
				}
				return DealDamage{Source: ctx.Source(), Target: c.Controller, Amount: 3}.Apply(ctx)
			}),
		),
	})
}
