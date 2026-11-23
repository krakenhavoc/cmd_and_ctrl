package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dark Favor — Enchantment — Aura {1}{B} (Magic 2012, #5573):
//
//	"Enchant creature
//	 When this Aura enters, you lose 1 life.
//	 Enchanted creature gets +3/+1."
//
// Chosen by Heliod's shape: PumpAttached for the bonus, and an enters
// trigger whose controller loses 1 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c1f4a440-638a-4e3e-8a3e-af4a4a73ec91",
		Name:         "Dark Favor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(3, 1)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Dark Favor — you lose 1 life", effects.Do(effects.YouLoseLife(1))),
		},
	})
}
