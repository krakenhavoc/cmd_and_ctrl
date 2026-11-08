package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dewdrop Cure — Sorcery {2}{W} (Bloomburrow, #4467):
//
//	"Gift a card (You may promise an opponent a gift as you cast this
//	 spell. If you do, they draw a card before its other effects.)
//	 Return up to two target creature cards each with mana value 2 or
//	 less from your graveyard to the battlefield. If the gift was
//	 promised, instead return up to three target creature cards each
//	 with mana value 2 or less from your graveyard to the battlefield."
//
// Gift (CR 702.174, ADR 0089) in Peerless Recycling's shape: the promise
// swaps the target clause for its three-card twin (CR 702.174m), and the
// opponent draws before the return. The return is the set-rule
// reanimations' body (returnLegalGraveyardTargetsToBattlefield): every
// card still legal comes back together, as one entry (CR 603.6a). The
// mana value is the card's in the graveyard (X is 0 there, CR 202.3e).
//
// No simplification.
func init() {
	upTo := func(label string, n int) *game.TargetSpec {
		return TargetCardInGraveyard(label, YouOwn(), Creature(), ManaValueLE(2)).WithCount(0, n)
	}
	Register(Spec{
		OracleID:     "07634620-a83a-4ca0-8a60-2a21368a2661",
		Name:         "Dewdrop Cure",
		Completeness: CompletenessFull,
		Targets:      upTo("up to two target creature cards each with mana value 2 or less from your graveyard", 2),
		Gift: GiftACard().Instead(
			upTo("up to three target creature cards each with mana value 2 or less from your graveyard", 3)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return returnLegalGraveyardTargetsToBattlefield(ctx)
		},
	})
}
