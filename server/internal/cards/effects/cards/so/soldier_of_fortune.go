package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soldier of Fortune — Creature — Human Mercenary {R}, 1/1 (Alliances,
// #5634):
//
//	"{R}, {T}: Target player shuffles their library."
//
// The {T} keeps it off the turn it arrives (CR 302.6). The shuffle
// randomizes the target player's library (CR 701.24a); a player who has
// left the game by the time it resolves is not a legal target, and the
// ability does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8b2367b3-3a8d-442f-8d58-6e4c03fa5a46",
		Name:         "Soldier of Fortune",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, {T}: Target player shuffles their library.",
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.S82OldB547TargetPlayerShufflesTheirLibrary,
		}},
	})
}
