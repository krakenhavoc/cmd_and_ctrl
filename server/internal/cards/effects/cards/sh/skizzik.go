package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skizzik — Creature — Elemental {3}{R}, 5/3 (Dominaria):
//
//	"Kicker {R} (You may pay an additional {R} as you cast this spell.)
//	 Trample, haste
//	 At the beginning of the end step, if this creature wasn't kicked, sacrifice
//	 it."
//
// Ball Lightning's sacrifice at the beginning of every end step (CR 513.1), an
// intervening "if" (CR 603.4) read off the record the cast left on the
// permanent: a Skizzik that was never cast (reanimated, flickered) was not
// kicked either, and is sacrificed. A Skizzik some other player has gained
// control of is not sacrificed by its old controller's trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0c77d37e-791c-415e-b369-55e93147ec34",
		Name:            "Skizzik",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		OptionalCosts:   []game.AdditionalCost{effects.Kicker("{R}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, effects.AllOf(effects.AnyPlayer, effects.OldB208ThisWasntKicked),
				"Skizzik — it wasn't kicked, sacrifice it", effects.OldB208SacrificeThisIfNotKicked),
		},
	})
}
