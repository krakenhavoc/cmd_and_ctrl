package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Silent Skimmer — Creature — Eldrazi Drone {3}{B}, 0/4 (Battle for Zendikar,
// #5547):
//
//	"Devoid (This card has no color.)
//	 Flying
//	 Whenever this creature attacks, defending player loses 2 life."
//
// Keeper of Tresserhorn's defending player: the player attacked, or the
// controller of the planeswalker or battle attacked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "92198967-e89b-4213-a0fa-71361c8420dc",
		Name:            "Silent Skimmer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Silent Skimmer — defending player loses 2 life",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					defender := effects.DefendingPlayerOfAttack(g, item.Trigger.Event)
					if !effects.DefendingPlayerStillIn(g, defender) {
						return nil
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, defender, -2)
				}),
		},
	})
}
