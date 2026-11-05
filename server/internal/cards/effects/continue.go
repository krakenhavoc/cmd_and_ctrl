package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Continue? — Instant {1}{W} (TMNT Commander):
//
//	"Choose up to four target creature cards in your graveyard that were
//	 put there from the battlefield this turn. Return them to the
//	 battlefield."
//
// The clause is Samwise the Stouthearted's predicate on a creature
// filter with Behold the Sinister Six!'s count and return. "Put there
// from the battlefield this turn" reads each card's newest arrival in
// the graveyard (putIntoGraveyardFromBattlefieldThisTurn): a creature
// that died, went back to a hand and was then discarded is a new object
// that came from the hand (CR 400.7), and it is not a legal target. A
// token is not a card (CR 108.2).
//
// On resolution (CR 608.2b) a target that left the graveyard in
// response is skipped and the rest still return, together, as one entry
// (CR 603.6a), under their owner's control: the caster's, since the
// graveyard is "your graveyard".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "175c76c8-f36c-4948-aaa8-0208eac5f0d3",
		Name:         "Continue?",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("up to four target creature cards in your graveyard that were put there from the battlefield this turn",
			Creature(), YouOwn(), putThereFromTheBattlefieldThisTurn).WithCount(0, 4),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return returnLegalGraveyardTargetsToBattlefield(ctx)
		},
	})
}
