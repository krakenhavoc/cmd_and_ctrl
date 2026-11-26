package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Calculated Dismissal — Instant {2}{U} (Magic Origins, #5549):
//
//	"Counter target spell unless its controller pays {3}.
//	 Spell mastery — If there are two or more instant and/or sorcery cards in
//	 your graveyard, scry 2."
//
// Mana Leak's counter, then the scry when spell mastery holds, read as the
// spell resolves (it does not count itself). The scry is the caster's and
// happens whether or not the counter was paid for.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8e63065-9ebc-419d-b7af-5867fd81edff",
		Name:         "Calculated Dismissal",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 2},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return effects.CounterUnlessPaid{
						StackID:  t.ID,
						Cost:     "{3}",
						Question: "Calculated Dismissal — pay {3} or your spell is countered",
					}.Apply(ctx)
				}
				return nil
			}),
			effects.If{Cond: effects.OldB318SpellMastery(), Then: []effects.Applier{effects.Scry{N: 2}}},
		),
	})
}
