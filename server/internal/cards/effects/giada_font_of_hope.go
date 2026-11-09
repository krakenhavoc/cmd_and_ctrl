package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Giada, Font of Hope — Legendary Creature — Angel {1}{W}, 2/2:
//
//	"Flying, vigilance
//	 Each other Angel you control enters with an additional +1/+1
//	 counter on it for each Angel you already control.
//	 {T}: Add {W}. Spend this mana only to cast an Angel spell."
//
// The counter count is read as the Angel enters, off the battlefield
// (the entering creature is not on it yet), so Giada counts herself
// and an Angel that entered after her counts the same.
func init() {
	Register(Spec{
		OracleID:        "48e6d3d8-2f27-4017-acdd-40bce8cdbc02",
		Name:            "Giada, Font of Hope",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			Produced:     "{W}",
			Restrictions: []string{ManaRestrictCast, ManaRestrictSubtype("Angel")},
			Label:        "Add {W}. Spend this mana only to cast an Angel spell",
		}},
		Replacements: []game.ReplacementEffect{
			cmdrB1OtherAngelsEnterWithCountersForEachAngel("Giada, Font of Hope: each other Angel you control enters with an additional +1/+1 counter for each Angel you already control"),
		},
	})
}
