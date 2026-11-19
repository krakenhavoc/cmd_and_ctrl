package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fiend Artisan — Creature — Nightmare {B/G}{B/G}, 1/1:
//
//	"This creature gets +1/+1 for each creature card in your graveyard.
//	 {X}{B/G}, {T}, Sacrifice another creature: Search your library for a
//	 creature card with mana value X or less, put it onto the battlefield,
//	 then shuffle. Activate only as a sorcery."
//
// The first line is a layer 7c modification (it is not a characteristic-
// defining ability: the card prints 1/1), counting creature cards in the
// controller's graveyard live. The ability is Citanul Flute's X search put
// onto the battlefield with a sacrifice-another cost; X is announced with the
// mana cost (XMatters) and the sacrifice is paid with the rest of the costs,
// so the sacrificed creature is not counted until it is a card in the
// graveyard, which is after the cost is paid and the search is made.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "43b8456a-3333-4936-a09c-324327619c36",
		Name:         "Fiend Artisan",
		Completeness: CompletenessFull,
		XMatters:     true,
		Static: []game.StaticAbility{{
			Label:    "This creature gets +1/+1 for each creature card in your graveyard.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				p := g.PlayerByIDForEffect(source.Controller)
				if p == nil || p.Graveyard == nil {
					return
				}
				n := 0
				for _, gc := range p.Graveyard.Cards {
					if gc.IsCreature() {
						n++
					}
				}
				c.Power += n
				c.Toughness += n
			},
		}},
		Activated: []ActivatedAbility{{
			Label:        "{X}{B/G}, {T}, Sacrifice another creature: Search your library for a creature card with mana value X or less, put it onto the battlefield, then shuffle. Activate only as a sorcery.",
			Purpose:      game.Purpose{Tutors: 1},
			Cost:         Plus(ManaCost("{X}{B/G}"), TapCost(), SacrificeAnotherN(1, "another creature", Creature(), YouControl())),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := ctx.X()
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsCreature() && c.ManaValue() <= x },
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Shuffle:   true,
					Reason:    "Fiend Artisan — a creature card with mana value X or less",
				}.Apply(ctx)
			},
		}},
	})
}
