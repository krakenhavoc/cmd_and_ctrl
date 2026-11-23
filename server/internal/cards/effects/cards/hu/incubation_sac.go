package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "478cc55b-dd71-4adc-909d-c39d435587ee",
		Name:         "Incubation Sac",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.B10EntersWithCounters(effects.OilCounter, 3, "Incubation Sac: enters with three oil counters on it")},
		Activated: []effects.ActivatedAbility{{
			Label:        "{4}, {T}, Remove an oil counter from this artifact: Create a 3/3 colorless Phyrexian Golem artifact creature token. Activate only as a sorcery.",
			Purpose:      game.Purpose{Tokens: 1},
			Cost:         effects.Plus(effects.ManaCost("{4}"), effects.TapCost(), effects.RemoveCountersFromThis(effects.OilCounter, 1)),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 colorless Phyrexian Golem artifact"), N: 1}),
		}},
	})
}
