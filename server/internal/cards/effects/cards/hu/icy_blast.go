package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Icy Blast — Khans of Tarkir (S82 OLD b347, #5554):
//
//	Instant {X}{U}: Tap X target creatures. Ferocious — If you control a creature
//	with power 4 or greater, those creatures don't untap during their controllers'
//	next untap steps.
//
// Ferocious is read through YouControlPowerFourOrGreater (current power, layered). The freeze covers the creatures tapped here, each under its own
// controller's next untap step.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f49302c5-8510-4360-841c-a59f53f87e0b",
		Name:         "Icy Blast",
		Completeness: effects.CompletenessFull,
		Targets:      effects.OldB347XTargetCreatures(),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			if effects.FerociousYouControlPowerFourOrGreater(ctx) {
				return effects.TapAndFreeze{Targets: ids, Label: "Icy Blast — ferocious"}.Apply(ctx)
			}
			for _, id := range ids {
				if err := (effects.TapTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
