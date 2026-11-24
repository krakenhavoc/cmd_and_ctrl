package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magus of the Order — Creature — Human Wizard {2}{G}{G}, 3/3 (Commander
// Legends):
//
//	"{G}, {T}, Sacrifice this creature and another green creature: Search
//	 your library for a green creature card, put it onto the battlefield,
//	 then shuffle."
//
// Both sacrifices are part of the cost (CR 118.3), paid as the ability is
// activated: the Magus itself and one other green creature, so it can't pay
// with itself twice. "Green" is a colour word of the clause's text, read
// through the lens (OfColor). The search names a quality, so the controller
// may find nothing (CR 701.23b); either way the library is shuffled. The
// creature sacrificed with the Magus is not in the library, so it can't be
// the one found.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff922641-142a-4774-be3e-36a5183fce2b",
		Name:         "Magus of the Order",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}, Sacrifice this creature and another green creature: Search your library for a green creature card, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost(), effects.SacrificeThis(), effects.SacrificeAnotherN(1, "another green creature", effects.Creature(), effects.OfColor("G"))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				green := ctx.TextColor("G")
				return effects.SearchLibrary{
					Player:    ctx.Controller(),
					Predicate: func(c game.Card) bool { return c.IsCreature() && c.HasColor(green) },
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Shuffle:   true,
					Reason:    "Magus of the Order — a green creature card",
				}.Apply(ctx)
			},
		}},
	})
}
