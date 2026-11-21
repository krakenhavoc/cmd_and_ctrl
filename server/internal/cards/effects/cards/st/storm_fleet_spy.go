package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Storm Fleet Spy — Creature — Human Pirate {2}{U}, 2/2 (Foundations,
// #4465):
//
//	"Raid — When this creature enters, if you attacked this turn, draw
//	 a card."
//
// Fire Nation Raider's raid enters trigger with a draw for its Clue.
// Raid is an intervening "if" (CR 603.4): checked as it triggers and
// again as it resolves. Any attack declared this turn counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42081a57-b521-45c2-928d-0a0a4641818d",
		Name:         "Storm Fleet Spy",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.TmcB6YouAttackedThisTurn),
				"Storm Fleet Spy — Raid: draw a card",
				effects.Do(effects.If{Cond: effects.AvatarB11CondYouAttacked{}, Then: []effects.Applier{effects.DrawCards{N: 1}}})),
		},
	})
}
