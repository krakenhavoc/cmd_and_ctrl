package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skarrgan Skybreaker — Creature — Giant Shaman {4}{R}{R}{G}, 3/3
// (Guildpact, #5599):
//
//	"Bloodthirst 3 (If an opponent was dealt damage this turn, this
//	 creature enters with three +1/+1 counters on it.)
//	 {1}, Sacrifice this creature: It deals damage equal to its power to
//	 any target."
//
// The shared Bloodthirst entry replacement at three, and Voldaren
// Thrillseeker's fling of itself (the same printed ability): the damage
// is its power as it last existed on the battlefield (CR 608.2h), its
// counters included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a45d8d4-2a5e-4b5b-b430-fc2fd5fd82c0",
		Name:         "Skarrgan Skybreaker",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Bloodthirst(3, "Skarrgan Skybreaker")},
		Activated:    []effects.ActivatedAbility{effects.VoldarenThrillseekerSacrifice()},
	})
}
