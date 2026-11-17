package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Incubation Sac — Artifact {G} (Phyrexia: All Will Be One):
//
//	"This artifact enters with three oil counters on it.
//	 {4}, {T}, Remove an oil counter from this artifact: Create a 3/3
//	 colorless Phyrexian Golem artifact creature token. Activate only as a
//	 sorcery."
//
// The token is the table's Phyrexian Golem row (Blade Splicer's).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "478cc55b-dd71-4adc-909d-c39d435587ee",
		Name:         "Incubation Sac",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{b10EntersWithCounters(oilCounter, 3, "Incubation Sac: enters with three oil counters on it")},
		Activated: []ActivatedAbility{{
			Label:        "{4}, {T}, Remove an oil counter from this artifact: Create a 3/3 colorless Phyrexian Golem artifact creature token. Activate only as a sorcery.",
			Purpose:      game.Purpose{Tokens: 1},
			Cost:         Plus(ManaCost("{4}"), TapCost(), RemoveCountersFromThis(oilCounter, 1)),
			SorcerySpeed: true,
			Effect:       Do(CreateToken{Template: TokenCard("3/3 colorless Phyrexian Golem artifact"), N: 1}),
		}},
	})
}
