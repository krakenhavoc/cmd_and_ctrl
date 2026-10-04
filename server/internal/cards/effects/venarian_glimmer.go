package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Venarian Glimmer — Instant {X}{U}:
//
//	"Target player reveals their hand. You choose a nonland card with
//	 mana value X or less from it. That player discards that card."
//
// X is the value announced with the cast (CR 107.3a), read back as the
// spell resolves. Each card in the hand is measured as it is there: an
// {X} card's X is 0 (CR 202.3e), a split card's halves are combined,
// and a modal double-faced card is its front face. A hand with nothing
// that small is revealed and nothing is discarded (CR 609.3). Targeting
// yourself reveals your hand to everyone (the 2014-02-01 ruling).
//
// XMatters is left unset on purpose: X=0 is a real cast. The hand is
// still revealed, and a nonland card with mana value 0 can still be
// chosen (x_matters_guard_test.go's allowlist).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "336cc3ab-c1c7-4bf1-b1d7-dfc9f3c556ab",
		Name:         "Venarian Glimmer",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.X()
			return ChooseFromRevealedHand{
				Player: TargetedPlayer(ctx),
				Filter: And(Nonland(), ManaValueLE(x)),
				Label:  "nonland card with mana value " + strconv.Itoa(x) + " or less",
			}.Apply(ctx)
		},
	})
}
