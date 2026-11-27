package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thriss, Nantuko Primus — Legendary Creature — Insect Druid {5}{G}{G}, 5/5
// (Judgment):
//
//	"{G}, {T}: Target creature gets +5/+5 until end of turn."
//
// Baton of Courage's targeted pump with a mana-and-tap cost: the target is
// chosen as the ability is announced and gets the boost only if it is still
// a legal target as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "19ed8616-5d27-4f21-88a9-2af7cf5034ca",
		Name:         "Thriss, Nantuko Primus",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}: Target creature gets +5/+5 until end of turn.",
			Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 5, 5)), game.AnswerPump),
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.BoostTargetUntilEOT("Thriss, Nantuko Primus — +5/+5", 5, 5),
		}},
	})
}
