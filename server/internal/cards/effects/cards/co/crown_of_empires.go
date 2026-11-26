package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crown of Empires — Artifact {2} (Magic 2012, S82 older sets batch 451,
// #5573):
//
//	"{3}, {T}: Tap target creature. Gain control of that creature instead
//	 if you control artifacts named Scepter of Empires and Throne of
//	 Empires."
//
// "Instead" replaces the tap, so the stolen creature is not tapped; the
// condition is read as the ability resolves and the control change has no
// stated duration.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b1298ce8-b318-4d5b-83f2-e274726dfca5",
		Name:         "Crown of Empires",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Tap target creature. Gain control of that creature instead if you control artifacts named Scepter of Empires and Throne of Empires.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.S82OldB451CrownOfEmpiresAbility,
		}},
	})
}
