package cardsst

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sunbird Standard // Sunbird Effigy — a transforming artifact (#2709,
// ADR 0137 and its 2026-10-10 amendment):
//
//	Sunbird Standard — Artifact {3}
//	  "{T}: Add one mana of any color.
//	   Craft with one or more {5}"
//	Sunbird Effigy — Artifact Creature — Bird Construct, */*
//	  "Flying, vigilance, haste
//	   Sunbird Effigy's power and toughness are each equal to the number
//	   of colors among the exiled cards used to craft it.
//	   {T}: For each color among the exiled cards used to craft this
//	   creature, add one mana of that color."
//
// The bare "Craft with one or more" takes any other permanents you
// control and/or any cards in your graveyard, as many as you name
// (CraftWithOneOrMore with no card type). The Effigy reads CR
// 702.167c's link: its power and toughness are a characteristic-
// defining ability over the colours the materials still in exile have
// (five at most, per the ruling), and its mana ability is Bloom
// Tender's over the same colours.
//
// No simplification.
const SunbirdStandardOracleID = "c69742bf-f0f5-4216-a602-c21f9c5a6501"

func init() {
	effects.Register(effects.Spec{
		OracleID:     SunbirdStandardOracleID,
		Name:         "Sunbird Standard",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{
			effects.Craft("Craft with one or more {5}", "{5}", effects.CraftWithOneOrMore("")),
		},
	})

	effects.Register(effects.Spec{
		OracleID:        SunbirdStandardOracleID + "#1",
		Name:            "Sunbird Effigy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "haste"},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := len(effects.CraftMaterialColors(effects.CraftMaterialsOf(g, source)))
				c.Power, c.Toughness = n, n
			},
		}},
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			ProducedFunc: SunbirdEffigyProduced,
			Label:        "{T}: For each color among the exiled cards used to craft this creature, add one mana of that color",
		}},
	})
}

// sunbirdEffigyProduced is one mana of each colour among the Effigy's
// materials, WUBRG order; nothing when it was not crafted.
func SunbirdEffigyProduced(g *game.Game, _, source uuid.UUID) string {
	c, ok := g.LookupCardForEffect(source)
	if !ok {
		return ""
	}
	var b strings.Builder
	for _, col := range effects.CraftMaterialColors(effects.CraftMaterialsOf(g, &c)) {
		b.WriteString("{" + col + "}")
	}
	return b.String()
}
