package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nightsnare — Sorcery {3}{B}:
//
//	"Target opponent reveals their hand. You may choose a nonland card
//	 from it. If you do, that player discards that card. If you don't,
//	 that player discards two cards."
//
// The optional revealed-hand pick (#2115, ADR 0116's 2026-10-05
// amendment): the whole table sees the hand (CR 701.20a), and you may
// choose a nonland card or choose nothing. A chosen card is discarded.
// If you choose nothing, that player discards two cards of their own
// choice (CR 701.9b, its 2015-06-22 ruling), and that is also what
// happens when the hand has no nonland card to choose. The "if you
// don't" branch is a keyed continuation, so the open pick is a restore
// point.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9efd1bd-d82a-43ad-aba7-df172b9751ce",
		Name:         "Nightsnare",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ChooseFromRevealedHand{
				Player:   effects.TargetedPlayer(ctx),
				Filter:   effects.Nonland(),
				Label:    "nonland card",
				Optional: true,
				Then:     NightsnareIfYouDont,
			}.Apply(ctx)
		},
	})
}

// nightsnareIfYouDont is "If you don't, that player discards two
// cards."
var NightsnareIfYouDont = effects.RevealedPickThen("revealed-pick/nightsnare-discard-two",
	func(ctx *effects.Context, pick game.RevealedPick) error {
		if len(pick.Chosen) > 0 {
			return nil
		}
		ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player: pick.FromPlayer,
			Source: pick.Source,
			N:      2,
		})
		return nil
	})
