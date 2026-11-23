package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Katilda, Dawnhart Martyr // Katilda's Rising Dawn (#1855, ADR 0107
// §4) — a disturb card whose back face is an Aura.
//
// Front face, Legendary Creature — Spirit Warlock {1}{W}{W}, */*:
//
//	"Flying, lifelink, protection from Vampires
//	 Katilda's power and toughness are each equal to the number of
//	 permanents you control that are Spirits and/or enchantments.
//	 Disturb {3}{W}{W} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Legendary Enchantment — Aura:
//
//	"Enchant creature
//	 Enchanted creature has flying, lifelink, and protection from
//	 Vampires, and it gets +X/+X, where X is the number of permanents
//	 you control that are Spirits and/or enchantments.
//	 If Katilda's Rising Dawn would be put into a graveyard from
//	 anywhere, exile it instead."
//
// The front face's power and toughness are a characteristic-defining
// ability (CR 604.3, layer 7a). Katilda is a Spirit, so she counts
// herself. "Protection from Vampires" is the creature-type quality
// game/protection.go parses (CR 702.16). On the back face the grant
// is layer 6 and the pump layer 7c; "you" is the Aura's controller
// (CR 109.5), and the Aura is an enchantment, so it counts itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         KatildaOracleID,
		Name:             "Katilda, Dawnhart Martyr",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying", "lifelink", "protection from Vampires"},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Disturb("{3}{W}{W}")},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.SpiritsAndEnchantmentsYouControl(g, source)
				c.Power, c.Toughness = n, n
			},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     KatildaOracleID + "#1",
		Name:         "Katilda's Rising Dawn",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.GrantToAttached("flying", "lifelink", "protection from Vampires"),
			effects.PumpAttachedPer(1, 1, effects.SpiritsAndEnchantmentsYouControl),
		},
		Replacements: []game.ReplacementEffect{effects.DisturbedExile("Katilda's Rising Dawn")},
	})
}

const KatildaOracleID = "493b2820-c250-48bd-9f3c-e1b5639ee101"
