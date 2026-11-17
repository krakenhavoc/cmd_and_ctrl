package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Electrostatic Bolt — Instant {R} (Mirrodin, #5606):
//
//	"Electrostatic Bolt deals 2 damage to target creature. If it's an
//	 artifact creature, Electrostatic Bolt deals 4 damage to it instead."
//
// Steer Clear's two amounts, with the test made on the target itself as
// the spell resolves (CR 608.2h): the creature's types after every
// layer, so an artifact creature that lost the artifact type in
// response takes 2 and an animated artifact takes 4. A target that is
// no longer legal leaves the spell with nothing to do (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6fc26faa-94a8-433f-8863-126c2d2e73b0",
		Name:         "Electrostatic Bolt",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			amount := 2
			if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsArtifact() {
				amount = 4
			}
			return DealDamage{Source: ctx.Source(), Target: id, Amount: amount}.Apply(ctx)
		},
	})
}
