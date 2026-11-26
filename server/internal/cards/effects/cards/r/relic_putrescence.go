package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Relic Putrescence — Enchantment — Aura {2}{B} (Scars of Mirrodin, S82 OLD
// batch 472, #5577):
//
//	"Enchant artifact
//	 Whenever enchanted artifact becomes tapped, its controller gets a poison
//	 counter."
//
// Contaminated Ground's "becomes tapped" trigger on the enchanted permanent
// (for mana or any other reason, CR 701.26a), with a poison counter in place of
// the life loss: the counter goes to the artifact's controller as the trigger
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e6fd429-f925-4ff7-9c6e-34c2c9d15d04",
		Name:         "Relic Putrescence",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("enchant artifact", effects.Artifact()),
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventTapCard, game.EventAttack},
			AppliesTo: effects.EnchantedPermanentBecameTapped,
			Key:       "Relic Putrescence — its controller gets a poison counter",
			Effect:    effects.OldB472EnchantedPermanentsControllerGetsPoison,
		}},
	})
}
