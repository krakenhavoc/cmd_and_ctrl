package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silburlind Snapper — {5}{U} Creature — Turtle 6/6 (Shadows over Innistrad,
// S82 OLD b293):
//
//	"This creature can't attack unless you've cast a noncreature spell this
//	 turn."
//
// An attack-target restriction (CR 508.1c) on the creature itself, read live
// at declaration from the per-turn cast tally.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eb2fe2f7-60f7-44d3-8d4a-4e102012435d",
		Name:         "Silburlind Snapper",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.OldB293CantAttackUnlessYouHaveCastANoncreatureSpell()},
	})
}
