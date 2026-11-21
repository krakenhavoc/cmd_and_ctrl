package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Check for Traps — Sorcery {1}{B} (Adventures in the Forgotten Realms):
//
//	"Target opponent reveals their hand. You choose a nonland card from it.
//	 Exile that card. If an instant card or a card with flash is exiled this
//	 way, they lose 1 life. Otherwise, you lose 1 life."
//
// "Otherwise" includes exiling nothing (an empty or all-land hand), so the
// caster loses 1 life then too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "41b2ea88-b3b9-44f0-8886-7b2f37b22ebf",
		Name:         "Check for Traps",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ChooseFromRevealedHand{
				Player: effects.TargetedPlayer(ctx),
				Filter: effects.Nonland(),
				Label:  "nonland card",
				Exile:  true,
				Then:   checkForTrapsLifeLoss,
			}.Apply(ctx)
		},
	})
}

// checkForTrapsLifeLoss is the last sentence: the opponent loses 1 life if the
// exiled card is an instant or has flash, otherwise the caster does.
var checkForTrapsLifeLoss = effects.RevealedPickThen("revealed-pick/check-for-traps-life-loss",
	func(ctx *effects.Context, pick game.RevealedPick) error {
		victim := pick.Chooser
		if len(pick.Chosen) > 0 {
			c := pick.Chosen[0]
			if c.HasCardType("instant") || game.HasKeyword(&c, "flash") {
				victim = pick.FromPlayer
			}
		}
		return ctx.Game.ChangePlayerLifeForEffect(pick.Source, victim, -1)
	})
