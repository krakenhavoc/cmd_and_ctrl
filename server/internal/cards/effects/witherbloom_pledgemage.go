package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Witherbloom Pledgemage — Creature — Treefolk Warlock {3}{B/G}{B/G}, 5/5
// (Strixhaven: School of Mages, #5502):
//
//	"Magecraft — Whenever you cast or copy an instant or sorcery spell, you
//	 gain 1 life."
//
// The shared Magecraft trigger over a plain life gain.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f10039ae-2695-4b0a-82b2-82ee79a992c7",
		Name:         "Witherbloom Pledgemage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Magecraft("Witherbloom Pledgemage — you gain 1 life (magecraft)", Do(GainLife{Amount: 1})),
		},
	})
}
