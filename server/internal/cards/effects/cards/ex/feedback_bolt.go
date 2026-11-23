package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "2fc5ebd9-e4af-4e2b-a2cc-a13fe051c5cb",
		Name:         "Feedback Bolt",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayerOrPlaneswalker(),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := effects.StickerArtifactsControlledBy(ctx.Game, item.Controller)
			targets := ctx.LegalTargets()
			if len(targets) == 0 || n <= 0 {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: targets[0].ID, Amount: n}.Apply(ctx)
		},
	})
}
