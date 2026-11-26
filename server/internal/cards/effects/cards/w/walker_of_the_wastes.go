package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Walker of the Wastes — Creature — Eldrazi {4}{C}, 4/4 (Oath of the Gatewatch,
// #5545):
//
//	"Trample
//	 This creature gets +1/+1 for each land you control named Wastes."
//
// A layer 7c self pump counted on every recompute (Myr Adapter's shape): the
// lands its controller controls whose NAME is Wastes, whatever else they are
// (a land that has become a Wastes copy counts; a Forest does not).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fb818376-6e87-4aa3-a050-4b5f82942593",
		Name:            "Walker of the Wastes",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{effects.PumpSelfCreatureOrAttachedPer(1, 1, func(g *game.Game, source *game.Card) int {
			return effects.CountPermanentsYouControl(g, source, func(c *game.Card) bool {
				return c.IsLand() && game.CardNameMatches(*c, "Wastes")
			})
		})},
	})
}
