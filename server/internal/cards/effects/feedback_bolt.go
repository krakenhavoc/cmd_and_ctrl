package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feedback Bolt — Instant {4}{R} (Fifth Dawn, #5604):
//
//	"Feedback Bolt deals damage to target player or planeswalker equal
//	 to the number of artifacts you control."
//
// Lava Axe's "target player or planeswalker", with the damage counted as
// the spell resolves: the artifacts its caster controls then (Broodstar's
// count). With none, no damage is dealt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2fc5ebd9-e4af-4e2b-a2cc-a13fe051c5cb",
		Name:         "Feedback Bolt",
		Completeness: CompletenessFull,
		Targets:      targetPlayerOrPlaneswalker(),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := stickerArtifactsControlledBy(ctx.Game, item.Controller)
			for _, t := range ctx.LegalTargets() {
				if n <= 0 {
					return nil
				}
				return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
			}
			return nil
		},
	})
}
