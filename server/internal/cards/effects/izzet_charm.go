package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Izzet Charm — Instant for {U}{R}:
//
//	"Choose one —
//	 • Counter target noncreature spell unless its controller pays {2}.
//	 • Izzet Charm deals 2 damage to target creature.
//	 • Draw two cards, then discard two cards."
//
// S20 sub-PR 4. Mode 0 composes the S19 PayUnless prompt with
// CounterTarget: the targeted spell's controller gets the pay
// prompt, and declining counters the spell. Mode 2's discard is the
// S14 random-discard primitive — the "you choose" picker is still
// deferred.
func init() {
	Register(Spec{
		OracleID: "a07698f6-5ad5-49a3-9da2-f82d407f5cd7",
		Name:     "Izzet Charm",
		Modes: ChooseOne(
			Mode("Counter target noncreature spell unless its controller pays {2}.",
				TargetSpell("target noncreature spell", Noncreature())),
			Mode("Izzet Charm deals 2 damage to target creature.",
				TargetCreature("target creature")),
			Mode("Draw two cards, then discard two cards."),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			switch {
			case ctx.HasMode(0):
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				spellID := item.Targets[0].ID
				target := ctx.Game.StackItemForEffect(spellID)
				if target == nil {
					return nil
				}
				return PayUnless{
					Chooser:  target.Controller,
					Cost:     "{2}",
					Question: "Izzet Charm — pay {2} or your spell is countered?",
					OnDecline: func(ctx *Context) error {
						if ctx.Game.StackItemForEffect(spellID) == nil {
							return nil
						}
						return CounterTarget{StackID: spellID}.Apply(ctx)
					},
				}.Apply(ctx)
			case ctx.HasMode(1):
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: 2}.Apply(ctx)
			case ctx.HasMode(2):
				me := ctx.Controller()
				if err := (DrawCards{Player: me, N: 2}).Apply(ctx); err != nil {
					return err
				}
				return DiscardCards{Player: me, N: 2}.Apply(ctx)
			}
			return nil
		},
	})
}
