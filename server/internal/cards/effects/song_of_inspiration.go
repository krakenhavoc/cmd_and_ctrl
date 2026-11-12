package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Song of Inspiration — Instant {3}{G}{G} (Forgotten Realms Commander,
// #5681):
//
//	"Choose up to two target permanent cards in your graveyard. Roll a
//	 d20 and add the total mana value of those cards.
//	 1—14 | Return those cards to your hand.
//	 15+ | Return those cards to your hand. You gain life equal to their
//	 total mana value."
//
// The total mana value of the targets still legal as the spell resolves
// (CR 608.2b) is the roll's modifier (CR 706.2, RollDice.Modifier), read
// while they are in the graveyard, and the same number is the life the
// 15+ row gains. Both rows return those cards; with no target chosen the
// roll still happens, returns nothing and gains 0. With every chosen
// target gone the spell does not resolve and no die is rolled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bdb80d7b-672c-4e2a-b93b-9b96721b93f2",
		Name:         "Song of Inspiration",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("up to two target permanent cards in your graveyard",
			YouOwn(), Permanent()).WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			total := 0
			for _, id := range legalTargetCardIDs(ctx) {
				if c, ok := ctx.Game.LookupCardForEffect(id); ok {
					mv, _ := ctx.Game.ManaValueForEffect(c)
					total += mv
				}
			}
			return RollDice{Sides: 20, N: 1, Modifier: total, TailN: total, Then: songOfInspirationResult}.Apply(ctx)
		},
	})
}

// songOfInspirationResult is the results table, after the d20: the
// tail's N is the total mana value read as the spell resolved. An on-disk
// key: never renamed, never reused.
var songOfInspirationResult = CardTail("dice/song-of-inspiration-result", func(ctx *Context, o game.TailOutcome) error {
	result := DiceResult(o)
	if len(o.Rolls) == 0 || result < 1 {
		return nil
	}
	if err := returnLegalGraveyardTargetsToHand(ctx); err != nil {
		return err
	}
	if result < 15 {
		return nil
	}
	return GainLife{Player: ctx.Controller(), Amount: o.Tail.N}.Apply(ctx)
})
