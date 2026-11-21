package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Banishing Slash — Sorcery {W}{W} (Kamigawa: Neon Dynasty, #5493):
//
//	"Destroy up to one target artifact, enchantment, or tapped creature. Then
//	 if you control an artifact and an enchantment, create a 2/2 white Samurai
//	 creature token with vigilance."
//
// "Then" reads the board after the destroy, so destroying your own last
// artifact (or enchantment) turns the token off. With no target the clause
// is still asked, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2800830c-73c6-4097-be7b-d698c6813c38",
		Name:         "Banishing Slash",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("up to one target artifact, enchantment, or tapped creature", effects.Or(effects.Artifact(), effects.Enchantment(), effects.HobB3TappedCreature())).WithCount(0, 1),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			if !effects.NeoB08YouControlAnArtifactAndAnEnchantment(ctx.Game, ctx.Controller()) {
				return nil
			}
			return effects.CreateToken{Template: effects.NeoB08SamuraiToken(), N: 1}.Apply(ctx)
		},
	})
}
