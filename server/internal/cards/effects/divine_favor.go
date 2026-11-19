package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Divine Favor — Enchantment — Aura {1}{W} (Foundations Jumpstart):
//
//	"Enchant creature
//	 When this Aura enters, you gain 3 life.
//	 Enchanted creature gets +1/+3."
//
// The life gain is a real trigger on the Aura's own entry, so it can be
// responded to and still resolves if the Aura is gone (CR 603.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bffa18c3-642c-4239-b425-238599007976",
		Name:         "Divine Favor",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{PumpAttached(1, 3)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Divine Favor — you gain 3 life", Do(GainLife{Amount: 3})),
		},
	})
}
