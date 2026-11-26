package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Transgress the Mind — Sorcery {1}{B} (Battle for Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 Target player reveals their hand. You choose a card from it with mana
//	 value 3 or greater and exile that card."
//
// The revealed-hand pick (ADR 0116) filtered to mana value 3 or greater, and
// exiling instead of discarding (Appetite for Brains): not a discard, so no
// madness and no discard triggers. With no such card nothing is exiled, but the
// hand is still revealed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9a9bc316-9652-46d9-9448-57b8902f5d6f",
		Name:            "Transgress the Mind",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Targets:         effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ChooseFromRevealedHand{
				Player: effects.TargetedPlayer(ctx),
				Filter: effects.ManaValueGE(3),
				Label:  "card with mana value 3 or greater",
				Exile:  true,
			}.Apply(ctx)
		},
	})
}
