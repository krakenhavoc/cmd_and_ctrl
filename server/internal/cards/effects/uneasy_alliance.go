package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Uneasy Alliance — {1}{W} Enchantment — Aura (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"Enchant creature
//	 Enchanted creature can't attack or block.
//	 {5}, Sacrifice this Aura: Exile enchanted creature. You create a 1/1
//	 black Ninja creature token. Activate only as a sorcery."
//
// Spiral into Solitude's Aura with a plainer cost and a Ninja after the
// exile. "Enchanted creature" is the creature the Aura was attached to as
// it was sacrificed (exileEnchantedCreature), and the ability does not
// target. The Ninja is the activator's whether or not the exile found
// anything.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "989b93fe-cad0-4890-a228-8a20ca666f10",
		Name:         "Uneasy Alliance",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			RestrictAttached(game.CantAttackOrBlock),
		},
		Activated: []ActivatedAbility{{
			Label:        "{5}, Sacrifice this Aura: Exile enchanted creature. You create a 1/1 black Ninja creature token. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{5}"), SacrificeThis()),
			SorcerySpeed: true,
			Effect:       uneasyAllianceExileThenNinja,
		}},
	})
}

// uneasyAllianceExileThenNinja exiles the creature the Aura enchanted,
// then makes the activator's Ninja.
func uneasyAllianceExileThenNinja(g *game.Game, item *game.StackItem) error {
	if err := exileEnchantedCreature(g, item); err != nil {
		return err
	}
	return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 black Ninja"), N: 1}.Apply(NewContext(g, item))
}
