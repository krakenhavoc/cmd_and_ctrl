package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pick the Brain — Sorcery {2}{B} (Shadows over Innistrad, #5543):
//
//	"Target opponent reveals their hand. You choose a nonland card from it
//	 and exile that card.
//	 Delirium — If there are four or more card types among cards in your
//	 graveyard, search that player's graveyard, hand, and library for any
//	 number of cards with the same name as the exiled card, exile those
//	 cards, then that player shuffles."
//
// Soul Search's reveal-and-choose (ChooseFromRevealedHand with Exile, ADR
// 0116), then the extraction family's search (Invasive Surgery's, "any
// number") with the name of the card as it was in the hand. The graveyard
// is read once the card has moved (the spell itself is still on the stack,
// so it does not count). A hand with no nonland card chooses nothing, so
// there is no name to search for and no search.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bef3fca4-7290-416c-a100-1b7abbe7886e",
		Name:         "Pick the Brain",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ChooseFromRevealedHand{
				Player: effects.TargetedPlayer(ctx),
				Filter: effects.Nonland(),
				Label:  "nonland card",
				Exile:  true,
				Then:   pickTheBrainSearch,
			}.Apply(ctx)
		},
	})
}

var pickTheBrainSearch = effects.RevealedPickThen("revealed-pick/pick-the-brain-search",
	func(ctx *effects.Context, pick game.RevealedPick) error {
		if len(pick.Chosen) == 0 || !effects.DrcDeliriumMet(ctx.Game, ctx.Controller()) {
			return nil
		}
		return effects.ExtractionSearch(ctx, "Pick the Brain", pick.FromPlayer, effects.ExtractionNames(pick.Chosen[0]), false)
	})
