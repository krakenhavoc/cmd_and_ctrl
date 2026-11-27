package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Triclopean Sight — Enchantment — Aura {1}{W} (Lorwyn, S82 OLD, #5593):
//
//	"Flash
//	 Enchant creature
//	 When this Aura enters, untap enchanted creature.
//	 Enchanted creature gets +1/+1 and has vigilance."
//
// Airtight Alibi's flash and untap-on-entry around a +1/+1 and a vigilance
// grant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6441456a-cc20-48bb-95e9-c981fea8ea98",
		Name:            "Triclopean Sight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("vigilance"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Triclopean Sight — untap enchanted creature", effects.UntapAttachedHost),
		},
	})
}
