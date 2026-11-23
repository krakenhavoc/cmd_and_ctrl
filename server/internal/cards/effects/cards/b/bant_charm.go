package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bant Charm — Instant {G}{W}{U}:
//
//	"Choose one —
//	 • Destroy target artifact.
//	 • Put target creature on the bottom of its owner's library.
//	 • Counter target instant spell."
//
// Mode 1 puts the creature on the bottom of its OWNER's library (not
// the controller's), so a stolen creature goes home. Mode 2 only
// targets an instant SPELL — a sorcery or an ability is not a legal
// target — and the clause is checked at announce (CR 601.2c) and again
// at resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "324889d6-c857-41ed-bb60-408809fc9964",
		Name:         "Bant Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Destroy target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact()),
				effects.DestroyTheModesTarget),
			effects.ModeDoing("Put target creature on the bottom of its owner's library.",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return ctx.Game.TuckToLibraryForEffect(t.ID, true)
				}),
			effects.ModeDoing("Counter target instant spell.",
				effects.TargetSpell("target instant spell", effects.Instant()),
				effects.CounterTheModesTarget),
		),
	})
}
