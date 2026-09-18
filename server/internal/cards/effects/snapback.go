package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Snapback — Instant {1}{U} (EDHREC rank 4345):
//
//	"You may exile a blue card from your hand rather than pay this
//	 spell's mana cost.
//	 Return target creature to its owner's hand."
//
// A free Unsummon at the cost of a card, which is the whole Prophecy
// pitch cycle's deal. In Commander it is the answer a tapped-out blue
// deck keeps for the turn somebody's commander gets suited up: the
// mana is the resource you do not have, and the card is the one you
// do.
//
// The pitch is an ALTERNATIVE COST (CR 118.9), not a rider, and that
// distinction decides two things a resolution-time implementation
// would get backwards. The exile happens at announce with the Snapback
// already on the stack (CR 601.2a before 601.2h), so a countered
// Snapback still costs you the pitched card; and it is validated
// before anything is paid, so a hand with no blue card gets a rejected
// cast rather than a half-paid one.
//
// Pitch() takes a life component as well — Force of Will's 1 — and
// Snapback pays none, so it is zero. The card clause is
// CardInYourHand, which bakes in "your hand" rather than leaving it to
// the caller: a picker offering an opponent's cards would be both a
// rules bug and an information leak.
//
// Returning a creature to hand beats indestructible, regeneration and
// a counters-based commander at once, and it beats a token entirely
// (CR 111.8 — a bounced token ceases to exist).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e88de17c-086d-4e2a-b5b2-2f9f57ca7c0f",
		Name:         "Snapback",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			Pitch("Exile a blue card from your hand", 0,
				CardInYourHand("a blue card from your hand", OfColor("U")),
				"a blue card from your hand"),
		},
		Targets: TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BounceToHand{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
