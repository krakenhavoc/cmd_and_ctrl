package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lightform — {1}{W}{W} Enchantment:
//
//	"When this enchantment enters, it becomes an Aura with enchant creature.
//	 Manifest the top card of your library and attach this enchantment to it.
//	 Enchanted creature has flying and lifelink."
//
// The enters trigger is BecomeAnAuraManifestAndAttach (s82_old_b194_helpers.go),
// built on Necromancy's become-an-Aura and ADR 0154's swapped enchant clause;
// the grant is one layer-6 effect on the enchanted creature (CR 613.1f).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f2df375c-3741-409f-91d1-21bf632b2e1d",
		Name:         "Lightform",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Lightform — it becomes an Aura; manifest the top card of your library and attach it",
				effects.Do(effects.BecomeAnAuraManifestAndAttach{})),
		},
		Static: []game.StaticAbility{
			effects.GrantToAttached("flying", "lifelink"),
		},
	})
}
