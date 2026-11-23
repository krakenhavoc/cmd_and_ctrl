package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arno Dorian — Legendary Creature — Human Assassin {2}{B}{R}, 3/3
// (Assassin's Creed, #5948):
//
//	"Deathtouch
//	 Other Assassins you control get +2/+0.
//	 Disguise {B}{R} (You may cast this card face down for {3} as a 2/2
//	 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)"
//
// Bladestitched Skaab's power-only TribalAnthem over other Assassins you
// control, and the Disguise alternative cost. Face down, Arno has no
// abilities (CR 708.2a), so the anthem is off until he is turned up.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "ae90daea-6644-4235-bca6-b94ee9925f63",
		Name:             "Arno Dorian",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"deathtouch"},
		AlternativeCosts: []game.AlternativeCost{effects.Disguise("{B}{R}")},
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Assassin"}, Others: true, YoursOnly: true}, 2, 0),
		},
	})
}
