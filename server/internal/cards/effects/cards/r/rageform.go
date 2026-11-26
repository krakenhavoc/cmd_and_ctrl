package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rageform — {2}{R}{R} Enchantment (Fate Reforged):
//
//	"When this enchantment enters, it becomes an Aura with enchant creature.
//	 Manifest the top card of your library and attach this enchantment to it.
//	 Enchanted creature has double strike."
//
// Cloudform's shape: BecomeAnAuraManifestAndAttach (s82_old_b194_helpers.go)
// and a layer-6 grant on the enchanted creature (CR 613.1f).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a108899-76b4-4f85-8a59-91838fc26f4a",
		Name:         "Rageform",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rageform — it becomes an Aura; manifest the top card of your library and attach it",
				effects.Do(effects.BecomeAnAuraManifestAndAttach{})),
		},
		Static: []game.StaticAbility{
			effects.GrantToAttached("double strike"),
		},
	})
}
