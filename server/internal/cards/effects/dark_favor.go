package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c1f4a440-638a-4e3e-8a3e-af4a4a73ec91",
		Name:         "Dark Favor",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{PumpAttached(3, 1)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dark Favor — you lose 1 life", Do(youLoseLife(1))),
		},
	})
}
