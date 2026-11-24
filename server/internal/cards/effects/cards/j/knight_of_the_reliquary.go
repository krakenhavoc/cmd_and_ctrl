package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Knight of the Reliquary — Creature — Human Knight {1}{G}{W}, 2/2 (Secret
// Lair Drop):
//
//	"This creature gets +1/+1 for each land card in your graveyard.
//	 {T}, Sacrifice a Forest or Plains: Search your library for a land card,
//	 put it onto the battlefield, then shuffle."
//
// The static is a layer 7c modify on the source alone, read live on every
// layer pass over the land cards (not lands only by type line: a card in a
// graveyard is judged by its printed types) in its controller's graveyard. The
// ability pays {T} and sacrifices a Forest or Plains (a land with that
// subtype, Knight is not one) as a cost, so the sacrificed land is already in
// the graveyard when the search resolves; the found land enters untapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "393e5331-4456-4ba3-b410-bb630bb053db",
		Name:         "Knight of the Reliquary",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1, LandsUntapped: 1},
		Static: []game.StaticAbility{{
			Label:    "This creature gets +1/+1 for each land card in your graveyard.",
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
					if gc.IsLand() {
						n++
					}
				}
				c.Power += n
				c.Toughness += n
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label: "{T}, Sacrifice a Forest or Plains: Search your library for a land card, put it onto the battlefield, then shuffle.",
			Cost: effects.Plus(effects.TapCost(), effects.SacrificeN(1, "a Forest or Plains",
				effects.Or(effects.HasSubtype("Forest"), effects.HasSubtype("Plains")))),
			Purpose: game.Purpose{Tutors: 1, Lands: 1, LandsUntapped: 1, Answers: game.AnswerSacOutlet},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsLand() },
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Shuffle:   true,
					Reason:    "Knight of the Reliquary — a land card",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
