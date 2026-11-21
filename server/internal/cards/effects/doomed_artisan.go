package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Doomed Artisan — Creature — Human Artificer {2}{W}, 1/1 (Commander 2019,
// #5517):
//
//	"Sculptures you control can't attack or block.
//	 At the beginning of your end step, create a colorless Sculpture
//	 artifact creature token with "This token's power and toughness are
//	 each equal to the number of Sculptures you control.""
//
// The restriction is Hazoret, Godseeker's layer 6 CantAttack|CantBlock
// static, over every Sculpture its controller controls (the subtype read
// after the layers, so a creature made a Sculpture is held too). It lasts
// only while the Artisan is on the battlefield; the Sculptures it made stay
// and may attack once it is gone, as the rulings print.
//
// The Sculpture's "power and toughness are each equal to the number of
// Sculptures you control" is an ability of the TOKEN (ADR 0083), so it lives
// on the token's own catalog template: a layer 7a characteristic-defining
// static that counts its controller's Sculptures, itself included, wherever
// the token goes and whoever controls it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "70c92f25-32d6-4cda-b961-cf2680bd59ab",
		Name:         "Doomed Artisan",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Sculptures you control can't attack or block.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.HasSubtype(source.TextSubtype("Sculpture"))
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack | game.CantBlock
			},
		}},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Doomed Artisan — create a Sculpture token",
				Do(CreateToken{Template: c19DoomedArtisanSculptureToken(), N: 1})),
		},
	})
}

// c19DoomedArtisanSculptureToken is the colorless Sculpture artifact
// creature token with "This token's power and toughness are each equal to
// the number of Sculptures you control."
func c19DoomedArtisanSculptureToken() game.Card {
	return TokenFromCatalog(printedDoomedArtisanSculptureToken)
}

// printedDoomedArtisanSculptureToken is the Sculpture as printed. Its slug
// is an on-disk identity: never renamed, never reused.
func printedDoomedArtisanSculptureToken() TokenTemplate {
	return TokenTemplate{
		Slug: "doomed-artisan-sculpture",
		Card: game.Card{
			Name:              "Sculpture",
			TypeLine:          "Token Artifact Creature — Sculpture",
			VariableToughness: true, // printed */*
		},
		Static: []game.StaticAbility{{
			Label:     "This token's power and toughness are each equal to the number of Sculptures you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := B26PermanentsOfSubtypeControlled(g, source.Controller, source.TextSubtype("Sculpture"))
				c.Power, c.Toughness = n, n
			},
		}},
		Text: "This token's power and toughness are each equal to the number of Sculptures you control.",
	}
}
