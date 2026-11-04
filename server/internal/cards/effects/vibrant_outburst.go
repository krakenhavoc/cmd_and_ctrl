package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vibrant Outburst — Instant {U}{R} (Secrets of Strixhaven, #4456):
//
//	"Vibrant Outburst deals 3 damage to any target. Tap up to one target
//	 creature."
//
// Two target clauses read by slot (Homesickness' shape): slot 0 is the any
// target that takes 3 damage, slot 1 is "up to one" creature to tap, which
// may be the same creature, any creature (not only an opponent's), and may be
// left empty. Each half is re-checked at resolution and skipped when its
// target has left or is no longer legal (CR 608.2b), so choosing no creature
// still deals the damage and a vanished damage target still taps the creature.
// The damage is dealt first and the tap second, in the printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a0cbb1ba-73b1-4418-bbe3-5d1a02007c08",
		Name:         "Vibrant Outburst",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetAny(),
			TargetCreature("up to one target creature").WithCount(0, 1),
		),
		Purpose: game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok {
				if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, t := range ctx.ClauseTargets(1) {
				if !ctx.IsTargetLegal(t) || !onBattlefield(ctx.Game, t.ID) {
					continue
				}
				if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
