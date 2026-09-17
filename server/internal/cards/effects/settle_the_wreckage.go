package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Settle the Wreckage — Instant {2}{W}{W} (EDHREC rank 3048):
//
//	"Exile all attacking creatures target player controls. That
//	 player may search their library for that many basic land
//	 cards, put those cards onto the battlefield tapped, then
//	 shuffle."
//
// The one-sided combat wipe with a consolation prize. The target is
// a player; the set exiled is computed at resolution from the
// combat state (AttackingTarget is stamped at declaration —
// Aetherize's read) and swept as one simultaneous event through
// ExileAllMatching, whose continuation carries the count into the
// search. The search is THAT player's: their library, their prompt,
// "may" honoured so they can decline the whole thing, and the lands
// enter tapped through the search's own tapped clause.
//
// "That many" is the number of creatures that actually reached EXILE
// (#866). An attacking commander whose owner takes CR 903.9's offer
// went to the command zone rather than to exile, so it buys no land —
// and because that offer is a prompt, the search now waits for the
// answer rather than running with the commander counted on the
// strength of the question having been asked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2a2ea189-b663-4f4a-bb23-ff7a4af25f71",
		Name:         "Settle the Wreckage",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				return b29ExileAttackersThenTheyFetchBasics(ctx, t.ID)
			}
			return nil
		},
	})
}
