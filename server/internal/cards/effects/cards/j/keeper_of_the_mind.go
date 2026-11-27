package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keeper of the Mind — Creature — Human Wizard {U}{U}, 1/2 (Exodus, S82 OLD batch 742, #5626):
//
//	"{U}, {T}: Choose target opponent who has at least two more cards in hand
//	 than you do as you activate this ability. Draw a card."
//
// The opponent is only a condition: the hand sizes are read as the ability is
// activated (CR 602.2b), when the target is chosen, and the draw is yours
// whatever happens to the hands after. So the clause carries
// StillResolves: an opponent whose hand has shrunk in response does not make
// the ability fizzle (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "45393fba-b1cc-4db7-b7e0-3851bf421a65",
		Name:         "Keeper of the Mind",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}: Choose target opponent who has at least two more cards in hand than you do as you activate this ability. Draw a card.",
			Purpose: game.Purpose{Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
			Targets: effects.S82OldB742TargetOpponentWithAtLeastTwoMoreCardsInHand(),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
