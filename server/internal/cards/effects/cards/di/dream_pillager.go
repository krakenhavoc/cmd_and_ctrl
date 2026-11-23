package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dream Pillager — Creature — Dragon {5}{R}{R}, 4/4:
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, exile that many
//	 cards from the top of your library. Until end of turn, you may cast
//	 spells from among those exiled cards."
//
// The number is the combat damage just dealt (the trigger's event amount).
// The cards come off the top of the controller's own library and the
// permission is cast-only (CR 601.2): a land among them can't be played and
// stays exiled. The window ends with the turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "504cc26d-3fd5-4391-9d30-f0614e5c8bbc",
		Name:            "Dream Pillager",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Dream Pillager — exile that many cards from the top of your library; until end of turn, you may cast spells from among them",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
						return nil
					}
					return effects.ExileTopWithPermission{
						From:     item.Controller,
						GrantTo:  item.Controller,
						N:        item.Trigger.Event.Amount,
						CastOnly: true,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
