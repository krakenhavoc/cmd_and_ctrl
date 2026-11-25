package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dust Stalker — Creature — Eldrazi {2}{B}{R}, 5/3 (Battle for Zendikar,
// #5547):
//
//	"Devoid (This card has no color.)
//	 Haste
//	 At the beginning of each end step, if you control no other colorless
//	 creatures, return this creature to its owner's hand."
//
// An intervening "if" (CR 603.4): it is checked as each end step begins and
// again as the trigger resolves. Every player's end step counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "69cd25d2-206c-4572-95e7-d66973c94358",
		Name:            "Dust Stalker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "haste"},
		Triggered: []game.TriggeredAbility{
			effects.OldB306IfYouControlNoOtherColorless(
				effects.On(game.EventBeginEndStep, effects.AnyPlayer,
					"Dust Stalker — return this creature to its owner's hand",
					func(g *game.Game, item *game.StackItem) error {
						if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
							return nil
						}
						return effects.BounceToHand{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
					})),
		},
	})
}
