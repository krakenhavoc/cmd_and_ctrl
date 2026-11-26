package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rumbling Aftershocks — Enchantment {4}{R} (Worldwake, S82 OLD batch 489,
// #5582):
//
//	"Whenever you cast a kicked spell, you may have this enchantment deal
//	 damage to any target equal to the number of times that spell was kicked."
//
// The trigger is built only for a spell cast with kicker or multikicker paid,
// and carries the number of times it was kicked (CR 702.33d), read from the
// spell's cast record. The damage is optional and targets any target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "851a2214-5f87-4e5b-b099-1c5c90784a97",
		Name:         "Rumbling Aftershocks",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.S82OldB489WheneverYouCastAKickedSpell(
				"Rumbling Aftershocks — damage to any target equal to the number of times that spell was kicked",
				effects.S82OldB489DamageToTargetEqualToTimesKicked), effects.TargetAny()),
				"Rumbling Aftershocks — deal damage to the target?"),
		},
	})
}
