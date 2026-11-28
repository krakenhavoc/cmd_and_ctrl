package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cowardice — Enchantment {3}{U}{U} (Eighth Edition, #5607):
//
//	"Whenever a creature becomes the target of a spell or ability, return
//	 that creature to its owner's hand. (It won't be affected by the spell or
//	 ability.)"
//
// The trigger goes on the stack above the spell or ability that targeted
// the creature, so the creature is gone when that resolves and the target is
// illegal (CR 608.2b). It names the creature as it was (CR 400.7).
//
// No simplification.
func init() {
	const label = "Cowardice — return that creature to its owner's hand"
	effects.Register(effects.Spec{
		OracleID:     "2692dd66-c200-4d0b-9427-30deaa2342ab",
		Name:         "Cowardice",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventBecomesTarget},
			AppliesTo: effects.S82OldB615TargetedCreature,
			Key:       label,
			Build:     effects.BuildNamingTheEventTarget(label),
			Effect:    effects.S82OldB615BounceTheNamedCreature,
		}},
	})
}
