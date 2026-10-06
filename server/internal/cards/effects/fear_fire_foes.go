package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fear, Fire, Foes! — Sorcery {X}{R}:
//
//	"Damage can't be prevented this turn. Fear, Fire, Foes! deals X
//	 damage to target creature and 1 damage to each other creature with
//	 the same controller."
//
// The turn grant (ADR 0107 §5) begins first. "The same controller" is
// the target's controller as the spell resolves, read before any
// damage is dealt. The two damage instructions are one sentence and
// simultaneous in the rules (CR 101.4); the engine deals them in turn,
// target first, which no card at the table can tell apart.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID: "35a0e085-df04-454d-b4d3-42db387bbad0",
		Name:     "Fear, Fire, Foes!",
		// ADR 0126 §6: 1 damage to each other creature the target's controller controls.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true, Partial: true}},
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (DamageCantBePreventedThisTurn{}).Apply(ctx); err != nil {
				return err
			}
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			t := targets[0]
			target, ok := ctx.Game.LookupCardForEffect(t.ID)
			if !ok {
				return nil
			}
			var others []game.Card
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.InstanceID != t.ID && c.IsCreature() && c.Controller == target.Controller {
					others = append(others, c)
				}
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
				if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: ctx.X()}).Apply(ctx); err != nil {
					return err
				}
				for _, c := range others {
					if err := (DealDamage{Source: item.SourceCardID, Target: c.InstanceID, Amount: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}
