package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thoughtpicker Witch — Creature — Human Wizard {B}, 1/1 (Ravnica: City of
// Guilds, S82 OLD batch 574, #5600):
//
//	"{1}, Sacrifice a creature: Look at the top two cards of target
//	 opponent's library, then exile one of them."
//
// The creature may be the Witch herself. Only the controller looks
// (CR 701.20) and picks the card to exile; the other stays on top
// (S82OldB574ThoughtpickerLook).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f78d21b-1959-4edc-9f28-d7ada539abef",
		Name:         "Thoughtpicker Witch",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice a creature: Look at the top two cards of target opponent's library, then exile one of them.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeN(1, "a creature", effects.Creature())),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect:  effects.S82OldB574ThoughtpickerLook,
		}},
	})
}
