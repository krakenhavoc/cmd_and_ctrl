package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Achilles Davenport — Legendary Creature — Human Assassin {2}{U}{B}, 3/3
// (Assassin's Creed, #5948):
//
//	"Freerunning {U}{B} (You may cast this spell for its freerunning cost
//	 if you dealt combat damage to a player this turn with an Assassin or
//	 commander.)
//	 Menace
//	 Other Assassins you control get +1/+1."
//
// Lord of the Accursed's TribalAnthem over Assassins you control, the
// lord excluded, read through the layered view; freerunning is the
// Brotherhood Ambushers alternative cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "b921a672-721e-4bd8-a10b-a6de8c13c7d8",
		Name:             "Achilles Davenport",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"menace"},
		AlternativeCosts: []game.AlternativeCost{effects.Freerunning("{U}{B}", "Assassin")},
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Assassin"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
