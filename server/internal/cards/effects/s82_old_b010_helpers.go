package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b010_helpers.go — shared pieces of the Secret Lair Drop batch 10
// cards (#5462).

// OldB10VoiceElementalToken is Voice of Resurgence's token: a green and
// white Elemental whose power and toughness are each equal to the number of
// creatures you control, a characteristic-defining ability (CR 604.3,
// CR 208.2a) in layer 7a that counts the token's own controller's creatures.
func OldB10VoiceElementalToken() game.Card { return TokenFromCatalog(printedOldB10VoiceElementalToken) }

func printedOldB10VoiceElementalToken() TokenTemplate {
	return TokenTemplate{
		Slug: "elemental-voice-of-resurgence",
		Card: game.Card{
			Name:     "Elemental",
			TypeLine: "Token Creature — Elemental",
			Colors:   []string{"G", "W"},
		},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			Label:    "This token's power and toughness are each equal to the number of creatures you control.",
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := B04CreaturesControlled(g, source.Controller)
				c.Power, c.Toughness = n, n
			},
		}},
		Text: "This token's power and toughness are each equal to the number of creatures you control.",
	}
}

// OldB10CountSubtype counts the creatures `player` controls with the creature
// type (through the layered view, so a changeling counts).
func OldB10CountSubtype(g *game.Game, player uuid.UUID, subtype string) int {
	return CountControlled(g, player, func(c game.Card) bool { return c.IsCreature() && c.HasSubtype(subtype) })
}
