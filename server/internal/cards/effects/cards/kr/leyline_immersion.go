package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leyline Immersion — Enchantment — Aura {3}{G} (March of the Machine:
// The Aftermath):
//
//	"Enchant legendary creature
//	 Enchanted creature has ward {2} and "{T}: Add five mana in any
//	 combination of colors. Spend this mana only to cast spells.""
//
// Two grants on the host. The ward is WardAttached (Lavaspur Boots'
// shape: the Aura carries the ward trigger, watching its host). The mana
// ability is a granted bundle (Find the Path's shape), so it is the
// creature's and its controller's: five independent any-colour slots
// (Baxter Building's AnyCombinationOfColors) carrying the CR 106.12
// restriction to casting spells.
//
// No simplification.
func init() {
	const grant = "leyline-immersion/five-mana"
	effects.Register(effects.Spec{
		OracleID:     "aa6c4079-72e1-4b2e-bd26-9399298c6105",
		Name:         "Leyline Immersion",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("enchant legendary creature", effects.Legendary()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key:  grant,
			Text: "{T}: Add five mana in any combination of colors. Spend this mana only to cast spells.",
			Mana: []effects.ManaAbility{{
				Cost:         effects.ManaAbilityCost{Tap: true},
				Produced:     effects.AnyCombinationOfColors(5),
				Restrictions: []string{effects.ManaRestrictCast},
				Label:        "Add five mana in any combination of colors. Spend this mana only to cast spells",
			}},
		}},
		Static: []game.StaticAbility{effects.GrantAbilities(effects.AttachedToSource, grant)},
		Triggered: []game.TriggeredAbility{
			effects.WardAttached(effects.WardMana("{2}"), "Leyline Immersion — enchanted creature has ward {2}"),
		},
	})
}
