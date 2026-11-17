package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "aa6c4079-72e1-4b2e-bd26-9399298c6105",
		Name:         "Leyline Immersion",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("enchant legendary creature", Legendary()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []AbilityGrant{{
			Key:  grant,
			Text: "{T}: Add five mana in any combination of colors. Spend this mana only to cast spells.",
			Mana: []ManaAbility{{
				Cost:         ManaAbilityCost{Tap: true},
				Produced:     AnyCombinationOfColors(5),
				Restrictions: []string{ManaRestrictCast},
				Label:        "Add five mana in any combination of colors. Spend this mana only to cast spells",
			}},
		}},
		Static: []game.StaticAbility{GrantAbilities(AttachedToSource, grant)},
		Triggered: []game.TriggeredAbility{
			WardAttached(WardMana("{2}"), "Leyline Immersion — enchanted creature has ward {2}"),
		},
	})
}
