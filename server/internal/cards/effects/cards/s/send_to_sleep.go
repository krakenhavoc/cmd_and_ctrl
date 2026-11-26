package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Send to Sleep — Instant {1}{U} (Magic Origins, #5549):
//
//	"Tap up to two target creatures.
//	 Spell mastery — If there are two or more instant and/or sorcery cards in
//	 your graveyard, those creatures don't untap during their controllers'
//	 next untap steps."
//
// Frost Breath's tap-and-freeze when spell mastery holds as the spell
// resolves, a plain tap when it does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e3fd1e9-7db6-40de-b1de-cd8cc9f60590",
		Name:         "Send to Sleep",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if effects.SpellMastery(ctx.Game, ctx.Controller()) {
				return effects.B751TapFreezeTargets(ctx, uuid.Nil, "Send to Sleep")
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.TapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
