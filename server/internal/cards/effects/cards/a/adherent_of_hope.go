package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Adherent of Hope — Creature — Human Soldier {1}{W}, 2/1 (Core Set 2021,
// #5511):
//
//	"At the beginning of combat on your turn, if you control a Basri
//	 planeswalker, put a +1/+1 counter on this creature."
//
// An intervening "if" (CR 603.4), checked as combat begins and again as the
// trigger resolves. "A Basri planeswalker" is the planeswalker with the Basri
// subtype, not a card merely named Basri.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b5f5400e-f341-40e3-b935-36d9cf983a20",
		Name:         "Adherent of Hope",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you control a Basri planeswalker",
				func(g *game.Game, you uuid.UUID) bool { return effects.B42ControlsPlaneswalkerNamed(g, you, "Basri") },
				effects.On(game.EventStepBegan, effects.StepBegan(game.StepBeginCombat, true),
					"Adherent of Hope — put a +1/+1 counter on this creature",
					effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}
