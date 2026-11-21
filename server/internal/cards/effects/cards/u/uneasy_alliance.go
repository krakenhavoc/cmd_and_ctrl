package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "989b93fe-cad0-4890-a228-8a20ca666f10",
		Name:         "Uneasy Alliance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{5}, Sacrifice this Aura: Exile enchanted creature. You create a 1/1 black Ninja creature token. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{5}"), effects.SacrificeThis()),
			SorcerySpeed: true,
			Effect:       uneasyAllianceExileThenNinja,
		}},
	})
}

// uneasyAllianceExileThenNinja exiles the creature the Aura enchanted,
// then makes the activator's Ninja.
func uneasyAllianceExileThenNinja(g *game.Game, item *game.StackItem) error {
	if err := effects.ExileEnchantedCreature(g, item); err != nil {
		return err
	}
	return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 black Ninja"), N: 1}.Apply(effects.NewContext(g, item))
}
