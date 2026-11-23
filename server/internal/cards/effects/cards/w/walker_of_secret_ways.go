package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Walker of Secret Ways — Creature — Human Ninja {2}{U}, 1/2 (Betrayers of
// Kamigawa, #5602):
//
//	"Ninjutsu {1}{U} ({1}{U}, Return an unblocked attacker you control
//	 to hand: Put this card onto the battlefield from your hand tapped
//	 and attacking.)
//	 Whenever this creature deals combat damage to a player, look at that
//	 player's hand.
//	 {1}{U}: Return target Ninja you control to its owner's hand. Activate
//	 only during your turn."
//
// The keyword is Ninjutsu("{1}{U}") (ninjutsu.go). The look is
// Silent-Blade Oni's LookAtHandForEffect on the player the damage event
// named. The bounce is gated by the DuringYourTurn activation condition
// (CR 602.5), so it is never offered on an opponent's turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "451a3f52-9810-4137-93b1-9594e7727c29",
		Name:         "Walker of Secret Ways",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer(
				"Walker of Secret Ways — look at that player's hand",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					g.LookAtHandForEffect(item.Controller, item.Trigger.Event.Target)
					return nil
				}),
		},
		Activated: []effects.ActivatedAbility{
			effects.Ninjutsu("{1}{U}"),
			{
				Label:     "{1}{U}: Return target Ninja you control to its owner's hand. Activate only during your turn.",
				Purpose:   game.Purpose{Answers: game.AnswerSurvive},
				Cost:      effects.ManaCost("{1}{U}"),
				Targets:   effects.TargetPermanent("target Ninja you control", effects.Subtype("Ninja"), effects.YouControl()),
				Condition: effects.DuringYourTurn(),
				Effect:    effects.BounceUpToOneTarget,
			},
		},
	})
}
