package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Riverwalk Technique — Instant {3}{U} (Tarkir: Dragonstorm, #4463):
//
//	"Choose one —
//	 • The owner of target nonland permanent puts it on their choice of the
//	   top or bottom of their library.
//	 • Counter target noncreature spell."
//
// The first mode is Misleading Motes' tuck on any nonland permanent: the
// owner, not the caster, picks the end of the library
// (ownerTucksFirstTargetTopOrBottom). The second is the plain counter.
// Either does nothing if its target is gone by resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a123aab8-3f3a-4187-abcb-21a99dcf13bb",
		Name:         "Riverwalk Technique",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("The owner of target nonland permanent puts it on their choice of the top or bottom of their library.",
				effects.TargetPermanent("target nonland permanent", effects.Nonland()),
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.OwnerTucksFirstTargetTopOrBottom(ctx, "Riverwalk Technique — put it on your choice of the top or bottom of your library")
				}),
			effects.ModeDoing("Counter target noncreature spell.",
				effects.TargetSpell("target noncreature spell", effects.Noncreature()), effects.CounterTheModesTarget),
		),
	})
}
