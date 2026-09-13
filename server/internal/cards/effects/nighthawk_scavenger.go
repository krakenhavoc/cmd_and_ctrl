package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nighthawk Scavenger — Creature — Vampire Rogue {1}{B}{B}, 1+*/3
// (EDHREC rank 1884):
//
//	"Flying, deathtouch, lifelink
//	 Nighthawk Scavenger's power is equal to 1 plus the number of
//	 card types among cards in your opponents' graveyards."
//
// Vampire Nighthawk that grows with the table's graveyards. The
// power is a characteristic-defining ability (CR 604.3) in layer 7a,
// Tarmogoyf's mechanism, reading the distinct printed card types
// across every opponent's graveyard as ONE set — a creature in two
// graveyards is one type. Three keywords on PrintedKeywords.
//
// Engine gap it shares with Tarmogoyf and Consuming Aberration, not
// the card's: the layer cache is invalidated by battlefield motion,
// counters, taps and turn changes, not by a card reaching a
// graveyard from a hand or a library, so a mill or a discard shows
// on the power at the next recompute rather than at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "379ed4ec-8f2e-448b-9ed2-a3181e2f877e",
		Name:            "Nighthawk Scavenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch", "lifelink"},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = 1 + b17CardTypesInOpponentsGraveyards(g, source.Controller)
			},
		}},
	})
}
