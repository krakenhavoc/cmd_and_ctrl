package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Void Stalker — Creature — Elemental {1}{U}, 2/1 (Magic 2013, S82 OLD, #5568):
//
//	"{2}{U}, {T}: Put this creature and target creature on top of their
//	 owners' libraries, then those players shuffle their libraries."
//
// Both creatures go on top as the ability resolves (Run Aground's tuck, CR 701.24),
// each still on the battlefield, then the owner of each that went shuffles
// (a player who owns both shuffles once). A target gone in response leaves
// Void Stalker's own half to resolve (CR 608.2b); Void Stalker gone leaves the
// target's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2313eef6-5414-462c-bf2f-f3283786735f",
		Name:         "Void Stalker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}, {T}: Put this creature and target creature on top of their owners' libraries, then those players shuffle their libraries.",
			Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.S82OldB429TuckThisAndTarget,
		}},
	})
}
