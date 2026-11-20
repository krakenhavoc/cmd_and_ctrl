package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "639926c1-c4b4-4a49-b125-0e1b584c5bb8",
		Name:         "Klauth's Will",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Breathe Flame — Klauth's Will deals X damage to each creature without flying.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.AfrB18DamageEachCreatureWithoutFlying(ctx)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}}),
			effects.ModeWithPurpose(effects.ModeDoing("Smash Relics — Destroy up to X target artifacts and/or enchantments.",
				effects.SocB10PestInfestationTargets(),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					for _, t := range ctx.ModeTargets(occ) {
						if !ctx.IsTargetLegal(t) {
							continue
						}
						if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}), game.Purpose{Answers: game.AnswerRemove}),
		).OrUpToIf(2, effects.YouControlACommander),
	})
}
