package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Revenant — Creature — Spirit {4}{B}, */* (Stronghold, #5627):
//
//	"Flying
//	 Revenant's power and toughness are each equal to the number of
//	 creature cards in your graveyard."
//
// Flying is a printed keyword. The size is Splinterfright's
// characteristic-defining ability (CR 604.3), applied in layer 7a
// (CR 613.4a) over the creature cards in its controller's graveyard as
// they are now, so it changes as cards arrive and leave. With none it is
// 0/0 and dies to the state-based check (CR 704.5f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "05fc62d3-9b3c-409a-a623-9f81f395a5a0",
		Name:            "Revenant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:     "Power and toughness are each equal to the number of creature cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := b11CreatureCardsInGraveyard(g, source.Controller)
				c.Power, c.Toughness = n, n
			},
		}},
	})
}
