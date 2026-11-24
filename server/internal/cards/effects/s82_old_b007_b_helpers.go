package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b007_b_helpers.go — shared pieces of Secret Lair Drop batch 007
// (S82, #5462): Ajani Goldmane's Avatar.

// OldB007AvatarToken is Ajani Goldmane's Avatar: a white Avatar creature token
// with "This token's power and toughness are each equal to your life total."
func OldB007AvatarToken() game.Card { return TokenFromCatalog(printedOldB007AvatarToken) }

// printedOldB007AvatarToken is the Avatar as printed. Its slug is an on-disk
// identity: never renamed, never reused. The characteristic-defining ability is
// a static the token always prints (CR 604.3, 613.4a); it reads the life total
// live, so it declares DependsOnLifeTotal.
func printedOldB007AvatarToken() TokenTemplate {
	return TokenTemplate{
		Slug: "avatar-white-life-total",
		Card: game.Card{
			Name:     "Avatar",
			TypeLine: "Token Creature — Avatar",
			Colors:   []string{"W"},
		},
		Static: []game.StaticAbility{{
			Layer:              game.Layer7PT,
			SubLayer:           game.SubLayer7A_CDA,
			Label:              "This token's power and toughness are each equal to your life total.",
			DependsOnLifeTotal: true,
			AppliesTo:          SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				life := 0
				if p := g.PlayerByIDForEffect(source.Controller); p != nil {
					life = p.Life
				}
				c.Power, c.Toughness = life, life
			},
		}},
		Text: "This token's power and toughness are each equal to your life total.",
	}
}
