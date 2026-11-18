package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2800830c-73c6-4097-be7b-d698c6813c38",
		Name:         "Banishing Slash",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("up to one target artifact, enchantment, or tapped creature", Or(Artifact(), Enchantment(), hobB3TappedCreature())).WithCount(0, 1),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			if !neoB08YouControlAnArtifactAndAnEnchantment(ctx.Game, ctx.Controller()) {
				return nil
			}
			return CreateToken{Template: neoB08SamuraiToken(), N: 1}.Apply(ctx)
		},
	})
}
