package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Twisted Experiment — Enchantment — Aura {1}{B} (Urza's Destiny,
// #5621):
//
//	"Enchant creature
//	 Enchanted creature gets +3/-1."
//
// Grave Servitude's +3/-1 without the color change: a layer 7c pump
// while the Aura is attached. On a creature with 1 toughness it leaves
// a 0-toughness creature, which goes to its owner's graveyard as a
// state-based action (CR 704.5f). Declared a benefit, as Grave
// Servitude is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4066d4df-d98f-44cd-bf25-ebb5e9d9ddeb",
		Name:         "Twisted Experiment",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(3, -1),
		},
	})
}
