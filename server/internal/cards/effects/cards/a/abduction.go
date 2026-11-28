package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Abduction — Enchantment — Aura {2}{U}{U} (Weatherlight, S82 OLD b774):
//
//	"Enchant creature
//	 When this Aura enters, untap enchanted creature.
//	 You control enchanted creature.
//	 When enchanted creature dies, return that card to the battlefield under
//	 its owner's control."
//
// Control Magic's layer-2 control change plus an untap on entering. The dies
// trigger looks back at the creature as it died (CR 603.10a); the card returns
// as a new object under its owner (CR 400.7), whoever controlled it, and the
// Aura is gone by then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d0e1904e-1a37-41f6-8582-b9ea794bb886",
		Name:         "Abduction",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.ControlAttachedBySource()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Abduction — untap enchanted creature", effects.UntapAttachedHost),
			effects.WhenEnchantedCreatureDies("Abduction — return that card to the battlefield under its owner's control",
				effects.OldB774ReturnTheEventCardUnderItsOwnersControl),
		},
	})
}
