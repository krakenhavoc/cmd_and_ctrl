package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Klauth's Will — Instant {X}{R}{R}{G} (Adventures in the Forgotten Realms
// Commander):
//
//	"Choose one. If you control a commander as you cast this spell, you may
//	 choose both instead.
//	 • Breathe Flame — Klauth's Will deals X damage to each creature without
//	   flying.
//	 • Smash Relics — Destroy up to X target artifacts and/or enchantments."
//
// Akroma's Will's conditional mode count (OrUpToIf(2, YouControlACommander),
// read as the spell is cast, CR 601.2b) over two X bullets. Breathe Flame is
// Earthquake's creature half: the creatures without flying, read through the
// layers, are snapshotted before the first point lands. Smash Relics takes up
// to X targets (CountFromX, UpToX, as Kozilek's Command's graveyard bullet):
// the announced X is the ceiling, and each target still legal is destroyed.
// With both chosen they run in printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "639926c1-c4b4-4a49-b125-0e1b584c5bb8",
		Name:         "Klauth's Will",
		XMatters:     true,
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Breathe Flame — Klauth's Will deals X damage to each creature without flying.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					x := ctx.X()
					if x <= 0 {
						return nil
					}
					return ctx.Game.DamageInstanceForEffect(func() error {
						return damageEachMatching(ctx, And(Creature(), WithoutKeyword("flying")), x)
					})
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}}),
			ModeWithPurpose(ModeDoing("Smash Relics — Destroy up to X target artifacts and/or enchantments.",
				afrB18SmashRelicsTargets(),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					for _, t := range ctx.ModeTargets(occ) {
						if !ctx.IsTargetLegal(t) {
							continue
						}
						if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}), game.Purpose{Answers: game.AnswerRemove}),
		).OrUpToIf(2, YouControlACommander),
	})
}
