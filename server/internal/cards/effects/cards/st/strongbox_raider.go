package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Strongbox Raider — Creature — Orc Pirate {2}{R}{R}, 5/2 (Foundations,
// #4465):
//
//	"Raid — When this creature enters, if you attacked this turn, exile
//	 the top two cards of your library. Choose one of them. Until the end
//	 of your next turn, you may play that card."
//
// Gorehorn Raider's raid enters trigger (an intervening "if", CR 603.4)
// with Heroes' Hangout's Date Night body: both cards are exiled, the
// controller chooses one, and only that card may be played (a land is
// played, a spell cast) until the end of their next turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ba5a2689-3578-4b1a-99d6-6b6d2356330e",
		Name:         "Strongbox Raider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.TmcB6YouAttackedThisTurn),
				"Strongbox Raider — Raid: exile the top two cards of your library, choose one of them, you may play it until the end of your next turn",
				effects.Do(effects.If{Cond: effects.AvatarB11CondYouAttacked{}, Then: []effects.Applier{
					effects.Instruction(effects.StdB15ExileTopTwoChooseOnePlay),
				}})),
		},
	})
}
