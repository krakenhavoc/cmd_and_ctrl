package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nature's Embrace — {2}{G} Enchantment — Aura (Innistrad: Crimson Vow,
// S82):
//
//	"Enchant creature or land
//	 As long as enchanted permanent is a creature, it gets +2/+2.
//	 As long as enchanted permanent is a land, it has "{T}: Add two mana
//	 of any one color.""
//
// The pump is Aether Meltdown's pumpEnchantedCreature (it applies only
// while the host is a creature, in layer 7c after any type change). The
// land half is Discreet Retreat's keyed mana grant (one colour pick for
// both mana, ProducedOneColor), given only while the host is a land (a
// creature that is not a land gets no mana ability). A land creature gets both.
//
// No simplification.
func init() {
	const grant = "natures-embrace/land-mana"
	effects.Register(effects.Spec{
		OracleID:     "9ab101c2-215b-4736-873a-4a1ef6666b6d",
		Name:         "Nature's Embrace",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("enchant creature or land", effects.Or(effects.Creature(), effects.Land())),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key:  grant,
			Text: "{T}: Add two mana of any one color.",
			Mana: []effects.ManaAbility{{
				Cost:           effects.ManaAbilityCost{Tap: true},
				ProducedAmount: effects.ProducedOneColor(func(*game.Game, uuid.UUID, uuid.UUID) int { return 2 }),
				Label:          "{T}: Add two mana of any one color",
			}},
		}},
		Static: []game.StaticAbility{
			effects.PumpEnchantedCreature(2, 2),
			func() game.StaticAbility {
				ab := effects.GrantAbilities(func(target *game.Card, g *game.Game, source *game.Card) bool {
					return effects.AttachedToSource(target, g, source) && target.IsLand()
				}, grant)
				ab.Label = "As long as enchanted permanent is a land, it has \"{T}: Add two mana of any one color.\""
				return ab
			}(),
		},
	})
}
