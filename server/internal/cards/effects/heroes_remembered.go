package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heroes Remembered — Sorcery {6}{W}{W}{W} (Planar Chaos, #5595):
//
//	"You gain 20 life.
//	 Suspend 10—{W} (Rather than cast this card from your hand, you may pay
//	 {W} and exile it with ten time counters on it. At the beginning of your
//	 upkeep, remove a time counter. When the last is removed, you may cast it
//	 without paying its mana cost.)"
//
// A life gain with the Suspend special action (CR 702.62a), cast from
// hand for its mana cost or from exile for free.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f30ffd2-9910-4fbd-96f9-d66f1da506e9",
		Name:         "Heroes Remembered",
		Completeness: CompletenessFull,
		SpecialActions: []game.SpecialAction{
			Suspend(10, "{W}"),
		},
		OnResolve: ResolveSteps(GainLife{Amount: 20}),
	})
}
