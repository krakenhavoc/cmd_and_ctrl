package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blinding Drone — Creature — Eldrazi Drone {1}{U}, 1/3 (Oath of the Gatewatch, #5545):
//
//	"Devoid (This card has no color.)
//	 {C}, {T}: Tap target creature. ({C} represents colorless mana.)"
//
// Devoid makes it colourless (CR 702.114a). The tap is Arashin
// Sunshield's row. {C} is colorless mana only, never generic (CR
// 107.4c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c924e240-02a0-46b5-b004-d5c42db277ec",
		Name:            "Blinding Drone",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Activated: []ActivatedAbility{{
			Label:   "{C}, {T}: Tap target creature.",
			Cost:    Plus(ManaCost("{C}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
