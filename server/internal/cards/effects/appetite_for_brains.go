package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Appetite for Brains — Sorcery {B}:
//
//	"Target opponent reveals their hand. You choose a card from it
//	 with mana value 4 or greater and exile that card."
//
// The revealed-hand pick with an exile (#2115, ADR 0116's 2026-10-05
// amendment): the whole table sees the hand (CR 701.20a), only a card
// with mana value 4 or greater may be chosen — X is 0 in a hand (CR
// 202.3e, its 2018-12-07 ruling) — and the card is exiled. Exiling it
// is not discarding it (CR 701.9a), so madness and "whenever a player
// discards" never see it. A hand with no such card is revealed and
// nothing is exiled (CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ffc16e0-1778-4aba-ade3-58146c800f5f",
		Name:         "Appetite for Brains",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ChooseFromRevealedHand{
				Player: TargetedPlayer(ctx),
				Filter: ManaValueGE(4),
				Label:  "card with mana value 4 or greater",
				Exile:  true,
			}.Apply(ctx)
		},
	})
}
