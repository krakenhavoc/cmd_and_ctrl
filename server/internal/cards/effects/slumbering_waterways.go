package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slumbering Waterways — Land:
//
//	"Slumbering Waterways enters the battlefield tapped.
//	 {T}: Add {G} or {U}.
//	 Flying, vigilance, trample"
//
// A tapped dual land whose last line prints three keywords. They are on a
// land, so they do nothing until something makes it a creature, and then
// they are what it has (the printed keywords are read through the layers).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "271264fb-965c-4fe3-baec-2217de20f8b5",
		Name:            "Slumbering Waterways",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "trample"},
		Replacements:    []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G|U}",
			Label:    "Add {G} or {U}",
		}},
	})
}
