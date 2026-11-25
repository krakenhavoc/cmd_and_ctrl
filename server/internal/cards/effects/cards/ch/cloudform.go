package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cloudform — {1}{U}{U} Enchantment:
//
//	"When this enchantment enters, it becomes an Aura with enchant creature.
//	 Manifest the top card of your library and attach this enchantment to it.
//	 Enchanted creature has flying and hexproof."
//
// The enters trigger is BecomeAnAuraManifestAndAttach (s82_old_b194_helpers.go),
// built on Necromancy's become-an-Aura and ADR 0154's swapped enchant clause;
// the grant is one layer-6 effect on the enchanted creature (CR 613.1f).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2b9384a0-db2d-49e2-b121-6b81a0d77722",
		Name:         "Cloudform",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cloudform — it becomes an Aura; manifest the top card of your library and attach it",
				effects.Do(effects.BecomeAnAuraManifestAndAttach{})),
		},
		Static: []game.StaticAbility{
			effects.GrantToAttached("flying", "hexproof"),
		},
	})
}
