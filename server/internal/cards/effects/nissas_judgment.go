package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nissa's Judgment — Sorcery {4}{G} (Ravnica: Clue Edition):
//
//	"Support 2. (Put a +1/+1 counter on each of up to two target creatures.)
//	 Choose up to one target creature an opponent controls. Each creature
//	 you control with a +1/+1 counter on it deals damage equal to its power
//	 to that creature."
//
// Two target clauses: the support picks (up to two, any creatures, CR 701.41a)
// and up to one creature an opponent controls. The counters go on first
// (CR 608.2c), so a freshly supported creature already deals its bigger
// power. Each creature is its own damage source (deathtouch and lifelink
// apply) and the damage is one simultaneous event; powers are read as the
// instruction runs. No chosen target, or one that is gone, deals nothing.
func init() {
	Register(Spec{
		OracleID:     "14247820-5ecb-40bd-8dee-0625cc745ec2",
		Name:         "Nissa's Judgment",
		Completeness: CompletenessFull,
		Targets: Clauses(
			SupportTargets(2, false),
			TargetCreature("up to one target creature an opponent controls", OpponentControls()).WithCount(0, 1),
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				for _, t := range ctx.ClauseTargets(0) {
					if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
						continue
					}
					if err := nothingIfGone(ctx.Game.AddCounterByForEffect(ctx.Controller(), t.ID, game.CounterPlusOne, 1)); err != nil {
						return err
					}
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				victim, ok := ctx.ClauseTarget(1)
				if !ok || victim.Kind != game.TargetCard {
					return nil
				}
				return ctx.Game.DamageInstanceForEffect(func() error {
					for _, c := range MatchingBattlefield(ctx, And(YouControl(), CreatureWithCounter(game.CounterPlusOne))) {
						if err := (DealDamage{Source: c.InstanceID, Target: victim.ID, Amount: c.CurrentPower()}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				})
			}),
		),
	})
}
