package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Garland, Knight of Cornelia // Chaos, the Endless (Final Fantasy, #4462)
// — a transforming legendary creature.
//
// Front face, Legendary Creature — Human Knight {B}{R}, 3/2:
//
//	"Whenever you cast a noncreature spell, surveil 1.
//	 {3}{B}{B}{R}{R}: Return this card from your graveyard to the
//	 battlefield transformed. Activate only as a sorcery."
//
// Back face, Chaos, the Endless — Legendary Creature — Demon, 5/5, flying
// (chaos_the_endless.go).
//
// The surveil is one Do step on a cast trigger. The graveyard ability
// returns the same card to the battlefield with its back face up (CR 712.14a,
// ReturnFromGraveyardTransformedForEffect), and a card that left the
// graveyard in response is left alone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     garlandKnightOfCorneliaOracleID,
		Name:         "Garland, Knight of Cornelia",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Garland, Knight of Cornelia — surveil 1", Do(Surveil{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:        "{3}{B}{B}{R}{R}: Return this card from your graveyard to the battlefield transformed. Activate only as a sorcery.",
			Cost:         ManaCost("{3}{B}{B}{R}{R}"),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
					return nil
				}
				return ReturnFromGraveyard{
					Target: item.SourceCardID, Dest: game.ZoneBattlefield, Transformed: true,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

// garlandKnightOfCorneliaOracleID is shared with the back face, which
// registers under it plus "#1" (game.CatalogKey).
const garlandKnightOfCorneliaOracleID = "75e930dd-32e9-4c39-81a7-d833a3976b8d"
