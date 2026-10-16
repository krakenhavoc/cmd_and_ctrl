package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hallowed Haunting — Enchantment {2}{W}{W} (EDHREC rank 3123):
//
//	"As long as you control seven or more enchantments, creatures
//	 you control have flying and vigilance.
//	 Whenever you cast an enchantment spell, create a white Spirit
//	 Cleric creature token with "This token's power and toughness
//	 are each equal to the number of Spirits you control.""
//
// The enchantress deck's army. The gate is a layer 6 grant over the
// controller's creatures that applies only while they control
// seven or more enchantments, counted per recompute with the
// Haunting itself included (b29CreaturesYouControlWithSevenEnchantments);
// the trigger is Sigil of the Empty Throne's condition
// (enchantmentSpellCastByYou) and makes a Spirit Cleric.
//
// The Spirit Cleric's printed "power and toughness are each equal to
// the number of Spirits you control" is an ability of the TOKEN, so it
// lives on the token's own template (ADR 0083): a layer 7a
// characteristic-defining static the token carries wherever it goes,
// counting its controller's Spirits (itself included). It survives the
// Haunting leaving, as printed. The CDA sets the toughness
// (Characteristic.PTDefined), so a Cleric never lingers as a stand-in
// 0/0.
func init() {
	Register(Spec{
		OracleID:     "e310ab58-180e-4840-bc8d-9f9b06f1b478",
		Name:         "Hallowed Haunting",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16GrantKeywords(b29CreaturesYouControlWithSevenEnchantments, "flying", "vigilance"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, enchantmentSpellCastByYou, "Hallowed Haunting — create a Spirit Cleric", Do(CreateToken{Template: spiritClericToken(), N: 1})),
		},
	})
}

// spiritClericToken is the white Spirit Cleric creature token with "This
// token's power and toughness are each equal to the number of Spirits you
// control."
func spiritClericToken() game.Card { return tokenFromCatalog(printedSpiritClericToken) }

func printedSpiritClericToken() tokenTemplate {
	return tokenTemplate{
		Slug: "spirit-cleric",
		Card: game.Card{
			Name:              "Spirit Cleric",
			TypeLine:          "Token Creature — Spirit Cleric",
			Colors:            []string{"W"},
			VariableToughness: true, // printed */*
		},
		Static: []game.StaticAbility{{
			Label:    "This token's power and toughness are each equal to the number of Spirits you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := b29SpiritsControlled(g, source.Controller, source.TextSubtype("Spirit"))
				c.Power = n
				c.Toughness = n
			},
		}},
		Text: "This token's power and toughness are each equal to the number of Spirits you control.",
	}
}
