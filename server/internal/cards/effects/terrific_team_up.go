package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Terrific Team-Up — Instant {3}{G} (Marvel Super Heroes Commander, #4455):
//
//	"This spell costs {2} less to cast if you control a permanent with mana
//	 value 4 or greater.
//	 One or two target creatures you control each get +1/+0 until end of
//	 turn. They each deal damage equal to their power to target creature an
//	 opponent controls."
//
// Two target clauses: slot 0 one or two creatures you control (the damage
// sources), slot 1 the recipient. The reduction is a self cost modifier over a
// board read (Truck Toss's YouControlA), judged when the cost is determined
// (CR 601.2f) with the spell itself on the stack, so it never counts itself.
// On resolution each source still legal gets +1/+0 (even if the recipient
// left), then the sources each deal damage equal to their current power to
// the recipient as one damage event, each with its own deathtouch and
// lifelink. A source that left the battlefield deals nothing and gets nothing
// (CR 608.2b); a recipient that left takes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "431c5b04-7387-4724-adfa-d1eb7874e79e",
		Name:         "Terrific Team-Up",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("one or two target creatures you control", YouControl()).WithCount(1, 2),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		SelfCostModifiers: []game.CostModifier{
			CostsLess(2, "This spell costs {2} less to cast if you control a permanent with mana value 4 or greater.",
				YouControlA(ManaValueGE(4))),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var sources []game.TargetRef
			for _, t := range ctx.ClauseTargets(0) {
				if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
					sources = append(sources, t)
				}
			}
			for _, s := range sources {
				if err := (BoostUntilEOT{Target: s.ID, Power: 1, Label: "Terrific Team-Up — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
			}
			// The damage reads each source's power as it now stands.
			ctx.Game.RecomputeLayersIfStaleLocked()
			victim, ok := ctx.ClauseTarget(1)
			if !ok || victim.Kind != game.TargetCard {
				return nil
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
				for _, s := range sources {
					c, ok := ctx.Game.LookupCardForEffect(s.ID)
					if !ok {
						continue
					}
					if err := (DealDamage{Source: s.ID, Target: victim.ID, Amount: c.CurrentPower()}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}
