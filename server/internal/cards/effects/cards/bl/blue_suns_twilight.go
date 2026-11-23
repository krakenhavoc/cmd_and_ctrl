package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blue Sun's Twilight — Sorcery {X}{U}{U}:
//
//	"Gain control of target creature with mana value X or less. If X is
//	 5 or more, create a token that's a copy of that creature."
//
// The target is bound to the announced X (WithManaValueAtMostX, re-checked
// on resolution, CR 608.2b). The control change has no stated duration
// (CR 611.2a). The copy is made after the control change, from the creature
// as it now is (CR 707.2), and is controlled by the caster.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "11e637b4-6ce0-4514-98d6-02a2d073f73b",
		Name:         "Blue Sun's Twilight",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets: effects.TargetCreature("target creature with mana value X or less").
			WithManaValueAtMostX(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.GainControl{
					Target:     t.ID,
					Controller: item.Controller,
					Duration:   game.IndefiniteDuration(),
					Label:      "Blue Sun's Twilight — gain control (no stated duration)",
				}).Apply(ctx); err != nil {
					return err
				}
				if ctx.X() >= 5 {
					if err := (effects.CreateTokenCopy{Controller: item.Controller, Copy: t.ID, N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
}
