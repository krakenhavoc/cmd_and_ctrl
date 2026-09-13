package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elvish Reclaimer — Creature — Elf Warrior {G}, 1/2 (EDHREC rank
// 2064):
//
//	"This creature gets +2/+2 as long as there are three or more land
//	 cards in your graveyard.
//	 {2}, {T}, Sacrifice a land: Search your library for a land card,
//	 put it onto the battlefield tapped, then shuffle."
//
// A one-mana Crop Rotation on legs that grows once it has been
// cropping. The size is a layer 7c modify keyed on
// b17LandCardsInGraveyard (Multani's read); the ability is a CR 602
// activation with three cost components — the mana, the tap (which
// waits out summoning sickness, CR 302.1) and b08SacrificeALand's
// picker — whose land search is the S22 chooser with the tapped
// flag, any land card at all.
//
// Engine gap it shares with Multani, not the card's: the layer cache
// is invalidated by battlefield motion, counters and turn changes,
// not by a land reaching the graveyard from a hand or a library, so
// a discarded third land shows on the Reclaimer's size at the next
// recompute. The sacrifice that usually gets it there is battlefield
// motion and recomputes at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "94f01534-7bb3-4b1e-8f77-e84156408686",
		Name:         "Elvish Reclaimer",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && b17LandCardsInGraveyard(g, source.Controller) >= 3
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 2
				c.Toughness += 2
			},
		}},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}, Sacrifice a land: Search your library for a land card, put it onto the battlefield tapped, then shuffle.",
			Cost:  Plus(ManaCost("{2}"), TapCost(), b08SacrificeALand()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:        item.Controller,
					Predicate:     func(c game.Card) bool { return c.IsLand() },
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Elvish Reclaimer — a land card, onto the battlefield tapped",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
