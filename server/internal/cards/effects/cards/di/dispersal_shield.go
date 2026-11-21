package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dispersal Shield — Instant {1}{U} (Scourge):
//
//	"Counter target spell if its mana value is less than or equal to the
//	 greatest mana value among permanents you control."
//
// Any spell is a legal target; the comparison is made as Dispersal Shield
// resolves, with the target's mana value as it sits on the stack ({X}
// counts as the value chosen, CR 202.3e) against scgB1GreatestManaValue.
// When the target's mana value is greater, the spell resolves and does
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aefb1463-76fb-419e-8cda-8c5bfb88a3d2",
		Name:         "Dispersal Shield",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 || effects.SpellManaValueForEffect(ctx.Game, targets[0].ID) > effects.ScgB1GreatestManaValue(ctx) {
				return nil
			}
			return effects.CounterTarget{StackID: targets[0].ID}.Apply(ctx)
		},
	})
}
