package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Total War — Enchantment {3}{R} (Ice Age, S82 OLD, #5636):
//
//	"Whenever a player attacks with one or more creatures, destroy all untapped non-Wall creatures
//	 that player controls that didn't attack, except for creatures the player hasn't controlled
//	 continuously since the beginning of the turn."
//
// One trigger per declaration (OncePerBatchPerAttacker), whoever attacks. As it resolves, that player's
// untapped, non-Wall creatures they controlled since the turn began and that were not declared as
// attackers are destroyed; a vigilant attacker is untapped but did attack, so it stays.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8266dd1c-e262-4467-a893-6b3b5eedb427",
		Name:         "Total War",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatchPerAttacker(effects.On(game.EventAttack, effects.AnyPlayer,
				"Total War — destroy that player's untapped non-Wall creatures that didn't attack", effects.S82OldB842TotalWarDestroy)),
		},
	})
}
