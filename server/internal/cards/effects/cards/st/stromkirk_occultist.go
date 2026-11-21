package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stromkirk Occultist — Creature — Vampire Horror {2}{R}, 3/2:
//
//	"Trample
//	 Whenever this creature deals combat damage to a player, exile the top
//	 card of your library. Until end of turn, you may play that card.
//	 Madness {1}{R}"
//
// The impulse exile is Breeches' (ExileTopWithPermission); trample is a
// printed keyword and madness the engine's (CR 702.35a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3bdd140b-1406-4d07-8f35-78f2ba92de64",
		Name:            "Stromkirk Occultist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Madness:         "{1}{R}",
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Stromkirk Occultist — exile the top card; you may play it this turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ExileTopCardYouMayPlayThisTurn(item, effects.NewContext(g, item), 0)
				}),
		},
	})
}
