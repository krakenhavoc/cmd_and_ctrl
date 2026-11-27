package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orcish Spy — Creature — Orc Rogue {R}, 1/1 (Eighth Edition, S82 OLD batch 612,
// #5607):
//
//	"{T}: Look at the top three cards of target player's library."
//
// Dimir Machinations' look (LookAtTopOfPlayersLibraryForEffect: only the
// controller learns the cards, and the library is not reordered or changed). A
// target player gone by resolution makes it do nothing (CR 608.2b). A tap
// ability, so summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0d058991-3751-4afe-80c9-79f0c6e10282",
		Name:         "Orcish Spy",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Look at the top three cards of target player's library.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				player, ok := effects.FirstLegalPlayerTarget(effects.NewContext(g, item))
				if !ok {
					return nil
				}
				g.LookAtTopOfPlayersLibraryForEffect(item.Controller, player, 3)
				return nil
			},
		}},
	})
}
