package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Display of Power — Instant {1}{R}{R} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"This spell can't be copied.
//	 Copy any number of target instant and/or sorcery spells. You may choose
//	 new targets for the copies."
//
// "Can't be copied" is the spec flag Choreographed Sparks sets (CR 707.10
// applies to copies of this spell only). The targets are any number (zero
// included) of instant or sorcery spells, whoever controls them; each one still
// on the stack as it resolves is copied under the caster's control with the
// CR 707.10c new-targets choice, one copy per spell, in the order the targets
// were announced.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "99140891-face-4015-aacd-1309e87d8f9f",
		Name:         "Display of Power",
		Completeness: effects.CompletenessFull,
		CantBeCopied: true,
		Targets:      effects.TargetInstantOrSorcerySpell("any number of target instant and/or sorcery spells").WithCount(0, 0),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer {
					continue
				}
				if err := (effects.CopySpell{StackID: t.ID, Controller: ctx.Controller(), ChooseNewTargets: true}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
