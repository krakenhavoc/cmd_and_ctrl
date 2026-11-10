package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// gnome_soldier_token.go — The Lost Caverns of Ixalan's Gnome Soldier
// (Thousand Moons Smithy, Barracks of the Thousand, #4470): "a white
// Gnome Soldier artifact creature token with 'This token's power and
// toughness are each equal to the number of artifacts and/or creatures you
// control.'"
//
// A catalog template, not a tokens_table.go row, because the
// characteristic-defining ability is a static the token always prints
// (CR 604.3, 613.4a). An artifact creature counts once however many of
// those types it has, and the token counts itself, so it is at least 1/1
// (the rulings). The count is live: it is re-read on every recompute.

// GnomeSoldierToken is the Lost Caverns Gnome Soldier.
func GnomeSoldierToken() game.Card { return tokenFromCatalog(printedGnomeSoldierToken) }

// printedGnomeSoldierToken is the Gnome Soldier as printed. Its slug is an
// on-disk identity: never renamed, never reused.
func printedGnomeSoldierToken() tokenTemplate {
	return tokenTemplate{
		Slug: "gnome-soldier",
		Card: game.Card{
			Name:     "Gnome Soldier",
			TypeLine: "Token Artifact Creature — Gnome Soldier",
			Colors:   []string{"W"},
		},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			Label:     "This token's power and toughness are each equal to the number of artifacts and/or creatures you control.",
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for _, p := range g.BattlefieldCardsForEffect() {
					if p.Controller == source.Controller && (p.IsArtifact() || p.IsCreature()) {
						n++
					}
				}
				c.Power, c.Toughness = n, n
			},
		}},
		Text: "This token's power and toughness are each equal to the number of artifacts and/or creatures you control.",
	}
}
