package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crashing Boars — Creature — Boar {3}{G}{G}, 4/4 (Exodus, S82 OLD batch 744, #5626):
//
//	"Whenever this creature attacks, defending player chooses an untapped
//	 creature they control. That creature blocks this creature this turn if
//	 able."
//
// A choice, not a target (so hexproof and protection don't matter), made by the
// defending player (the player attacked, or the controller of the planeswalker
// or protector of the battle attacked, CR 506.2) as the trigger resolves, among
// the untapped creatures they control; with none, nothing happens. The chosen
// creature gets Provoke's block requirement, pinned to this Boars object
// (BlocksAttackerUntilEOT); a requirement it can't obey (a flier it can't
// reach, menace) asks nothing (CR 509.1c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a4c59feb-e76a-4e08-b296-21c3e61cfc64",
		Name:         "Crashing Boars",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Crashing Boars — defending player chooses an untapped creature they control; it blocks this creature this turn if able", effects.S82OldB744CrashingBoarsAttack),
		},
	})
}
