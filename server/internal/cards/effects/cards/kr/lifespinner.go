package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lifespinner — Creature — Spirit {3}{G}, 3/3 (Betrayers of Kamigawa, #5602):
//
//	"{T}, Sacrifice three Spirits: Search your library for a legendary Spirit
//	 permanent card, put it onto the battlefield, then shuffle."
//
// The tap and the three sacrificed Spirits are costs paid as the ability is
// activated (CR 602.2b); the Lifespinner is itself a Spirit and may be one of
// the three (it is not "another"). The search is for a legendary Spirit
// permanent card, which goes onto the battlefield, and the library is
// shuffled even when nothing is found. "Spirit" is a word of the text (CR
// 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0fb26acd-891d-471d-98d8-73c6f940052b",
		Name:         "Lifespinner",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice three Spirits: Search your library for a legendary Spirit permanent card, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Tutors: 1, Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(3, "three Spirits", effects.Subtype("Spirit"))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				spirit := ctx.TextSubtype("Spirit")
				return effects.SearchLibrary{
					Player: item.Controller,
					Predicate: func(c game.Card) bool {
						return c.IsPermanent() && c.HasSupertype("Legendary") && c.HasSubtype(spirit)
					},
					Dest:    game.ZoneBattlefield,
					Limit:   1,
					Shuffle: true,
					Reason:  "Lifespinner — a legendary Spirit permanent card",
				}.Apply(ctx)
			},
		}},
	})
}
