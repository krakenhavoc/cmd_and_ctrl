package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ancestral Vengeance — {B}{B} Enchantment — Aura (Fate Reforged):
//
//	"Enchant creature
//	 When this Aura enters, put a +1/+1 counter on target creature you control.
//	 Enchanted creature gets -1/-1."
//
// The enters trigger is its own ability with its own target (CR 603.3d), chosen
// as it goes on the stack; the -1/-1 is a layer-7c effect on the enchanted
// creature (CR 613.4c), so a creature the Aura takes to 0 toughness dies to the
// state-based check (CR 704.5f).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "321c4586-d27e-45ae-ba1f-3efea334a697",
		Name:         "Ancestral Vengeance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.PumpAttached(-1, -1)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Ancestral Vengeance — put a +1/+1 counter on target creature you control",
				// Animal Sanctuary's body: a +1/+1 counter on the announced
				// permanent, if it is still a legal target.
				effects.B36CounterOnChosenAnimal),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
