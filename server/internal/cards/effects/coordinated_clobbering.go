package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coordinated Clobbering — Sorcery {G} (Duskmourn: House of Horror,
// #4466):
//
//	"Tap one or two target untapped creatures you control. They each deal
//	 damage equal to their power to target creature an opponent controls."
//
// Two target clauses (Bite Down's shape, #764): slot 0 is one or two
// untapped creatures you control, slot 1 the opponent's creature. The
// untapped requirement is part of the clause, so a tapped creature is
// refused at announce (CR 601.2c) and one that is tapped, gone or
// otherwise illegal as the spell begins to resolve is skipped (CR 608.2b):
// it is not tapped and deals no damage. If the opponent's creature is
// illegal, no damage is dealt at all, but the creatures are still tapped.
//
// The legal creatures are read once, before any is tapped (tapping would
// otherwise make each fail its own "untapped" clause), then each deals
// damage equal to its power as the spell resolves, together as one damage
// event; the sources are the creatures, so deathtouch and lifelink carry.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b2b65b17-1fb4-42d6-9c6a-376d3f18851b",
		Name:         "Coordinated Clobbering",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // a one-sided fight at a creature
		Targets: Clauses(
			TargetCreature("one or two target untapped creatures you control", YouControl(), Untapped()).WithCount(1, 2),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var sources []game.TargetRef
			for _, t := range ctx.ClauseTargets(0) {
				if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
					sources = append(sources, t)
				}
			}
			for _, t := range sources {
				if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			victim, ok := ctx.ClauseTarget(1)
			if !ok || victim.Kind != game.TargetCard {
				return nil
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
				for _, t := range sources {
					c, ok := ctx.Game.LookupCardForEffect(t.ID)
					if !ok {
						continue
					}
					if err := (DealDamage{Source: t.ID, Target: victim.ID, Amount: c.CurrentPower()}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}
