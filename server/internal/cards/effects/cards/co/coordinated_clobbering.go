package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b2b65b17-1fb4-42d6-9c6a-376d3f18851b",
		Name:         "Coordinated Clobbering",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // a one-sided fight at a creature
		Targets: effects.Clauses(
			effects.TargetCreature("one or two target untapped creatures you control", effects.YouControl(), effects.Untapped()).WithCount(1, 2),
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var sources []game.TargetRef
			for _, t := range ctx.ClauseTargets(0) {
				if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
					sources = append(sources, t)
				}
			}
			for _, t := range sources {
				if err := (effects.TapTarget{Target: t.ID}).Apply(ctx); err != nil {
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
					if err := (effects.DealDamage{Source: t.ID, Target: victim.ID, Amount: c.CurrentPower()}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}
