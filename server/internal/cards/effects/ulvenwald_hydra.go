package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ulvenwald Hydra — Creature — Hydra {4}{G}{G}, */* (EDHREC rank
// 2804):
//
//	"Reach
//	 Ulvenwald Hydra's power and toughness are each equal to the
//	 number of lands you control.
//	 When this creature enters, you may search your library for a land
//	 card, put it onto the battlefield tapped, then shuffle."
//
// The ramp deck's six-drop that is usually a 7/7 and fetches the
// eighth land itself. The size is a characteristic-defining ability
// (CR 604.3) in layer 7a — Lumra's mechanism, counting lands — and
// the ETB is Wight of the Reliquary's land search with "you may",
// which is the search prompt's decline: any land card, not just a
// basic, onto the battlefield tapped through the search's own clause.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "40b85f70-78b6-427a-9fb9-c3a72b8528ed",
		Name:            "Ulvenwald Hydra",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := b10LandsControlled(g, source.Controller)
				c.Power = n
				c.Toughness = n
			},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ulvenwald Hydra — you may search for a land, tapped", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:        item.Controller,
					Predicate:     func(c game.Card) bool { return c.IsLand() },
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Shuffle:       true,
					TappedOnEntry: true,
					Optional:      true,
					Reason:        "Ulvenwald Hydra — a land card, onto the battlefield tapped",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
