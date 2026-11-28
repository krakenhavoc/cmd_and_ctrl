package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ashnod's Cylix — Artifact {2} (Alliances, #5634):
//
//	"{3}, {T}: Target player looks at the top three cards of their library, puts
//	 one of them back on top of their library, then exiles the rest."
//
// The TARGET player looks and chooses (a private look, nobody else sees the
// cards), keeps one on top and exiles the others, face up. A library of fewer
// than three looks at what it has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fdf4e4c3-fe17-4098-9f76-e98f3f00bd3e",
		Name:         "Ashnod's Cylix",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Target player looks at the top three cards of their library, puts one of them back on top of their library, then exiles the rest.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.OldB815AshnodsCylix,
		}},
	})
}
