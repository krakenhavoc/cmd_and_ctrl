package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Starseer Mentor — Creature — Bat Warlock {3}{W}{B}, 3/5 (Bloomburrow,
// #4467):
//
//	"Flying, vigilance
//	 At the beginning of your end step, if you gained or lost life this
//	 turn, target opponent loses 3 life unless they sacrifice a nonland
//	 permanent of their choice or discard a card."
//
// The trigger condition is Starlit Soothsayer's intervening if
// (CR 603.4), asked as your end step begins and again as the trigger
// resolves. The effect is Thornplate Intimidator's: Torment of
// Hailfire's three-way punisher asked once, of the targeted opponent
// (punisherRepeat.onTargetPlayer). An opponent who has left by
// resolution is asked nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.RegisterPunisher(starseerMentorPunisher, "starseer-mentor")
	effects.Register(effects.Spec{
		OracleID:        "12a165c4-229b-458b-8fc7-c1b7cba35a8e",
		Name:            "Starseer Mentor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventBeginEndStep, effects.YourEndStepAndYouGainedOrLostLifeThisTurn,
				"Starseer Mentor — target opponent loses 3 life unless they sacrifice a nonland permanent or discard a card",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.YouGainedOrLostLifeThisTurn(g, item.Controller) {
						return nil
					}
					return starseerMentorPunisher.OnTargetPlayer(effects.NewContext(g, item))
				}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}

// starseerMentorPunisher is the printed punisher: 3 life a refusal.
var starseerMentorPunisher = effects.PunisherRepeat{Name: "Starseer Mentor", Life: 3}
