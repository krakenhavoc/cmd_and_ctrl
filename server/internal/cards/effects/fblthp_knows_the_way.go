package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fblthp, Knows the Way — Legendary Creature — Homunculus Scout
// {X}{G}{G}, */2:
//
//	"Domain — Fblthp's power is equal to the number of basic land types
//	 among lands you control.
//	 When Fblthp enters, search your library for up to X basic land
//	 cards with different names, reveal them, put them into your hand,
//	 then shuffle."
//
// The power is a layer 7a characteristic-defining ability (Gaea's
// Liege's shape) over the controller's lands' post-layer subtypes, so a
// land that gained a basic type counts; each of the five types counts
// once however many lands carry it. The enters trigger reads the X the
// spell was cast with inside Build (Wan Shi Tong's pattern, CR 107.3m),
// so the trigger can be responded to and the search still uses the
// announced X. "Different names" is a Validate over the picked set.
//
// Small gap, harmless: with X = 0 the search is skipped outright
// instead of searching for nothing, so the library is not shuffled.
// Nothing observable differs — a shuffle of an unrevealed library.
func init() {
	Register(Spec{
		OracleID:     "48b891b6-3175-4864-bb98-af131098c557",
		Name:         "Fblthp, Knows the Way",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = rfCreatureBBasicLandTypesAmong(g, source.Controller)
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Key:       "Fblthp, Knows the Way — search for X basic lands with different names",
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Fblthp, Knows the Way — search for X basic lands with different names")
				item.Params.Amount = source.CastX()
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := item.Params.Amount
				if x <= 0 {
					return nil
				}
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     x,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Fblthp, Knows the Way — up to X basic land cards with different names",
					Validate:  rfCreatureBDistinctNames,
					Source:    item.SourceCardID,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
