package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Structural Collapse — Sorcery {5}{R} (Gatecrash, S82 OLD #5565):
//
//	"Target player sacrifices an artifact and a land of their choice.
//	 Structural Collapse deals 2 damage to that player."
//
// Two edicts in printed order, each the player's own choice (CR 701.21a): an
// artifact, then a land, so an artifact land may pay one but not both. A
// player without either sacrifices what they can; the 2 damage is dealt either
// way. A target gone as the spell resolves makes it do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "de97ec82-d09d-440e-8c7d-a3f0fb552418",
		Name:         "Structural Collapse",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				ctx.Game.PlayerSacrificesForEffect(ctx.Source(), t.ID, effects.SacrificeSpec("an artifact", effects.Artifact()), "Structural Collapse — sacrifice an artifact")
				ctx.Game.PlayerSacrificesForEffect(ctx.Source(), t.ID, effects.SacrificeSpec("a land", effects.Land()), "Structural Collapse — sacrifice a land")
				return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2}.Apply(ctx)
			}
			return nil
		},
	})
}
