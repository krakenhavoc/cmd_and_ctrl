package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jandor's Saddlebags — Artifact {2} (Arabian Nights, #5641):
//
//	"{3}, {T}: Untap target creature."
//
// Thousand-Year Elixir's untap row at {3}: any creature, whoever
// controls it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3aa0e73f-ac88-47a5-9cc5-0c941a939eae",
		Name:         "Jandor's Saddlebags",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Untap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.UntapFirstLegalTarget,
		}},
	})
}
