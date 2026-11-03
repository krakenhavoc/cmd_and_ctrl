package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fix What's Broken — Sorcery {2}{W}{B} (Secrets of Strixhaven, #4456):
//
//	"As an additional cost to cast this spell, pay X life.
//	 Return each artifact and creature card with mana value X from your
//	 graveyard to the battlefield."
//
// The life is Toxic Deluge's additional cost (PayXLifeCost): X is announced
// with the spell, the life is paid as it is cast (CR 601.2b, 601.2h), and the spell
// has no {X} in its mana cost, so X is a number the spell reads on the stack
// (CR 202.3e) and the cards are judged by their printed mana value in the
// graveyard, where an {X} in a cost counts as 0. "Each artifact and creature
// card" is every card that is either, in your graveyard as the spell
// resolves; they are read before any moves and enter together, as one event
// (Rise of the Dark Realms), so each sees the others enter (CR 603.6a). With
// X = 0 the cards of mana value 0 return (tokens never are cards).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "7f8f46c0-889a-4c9a-9bfa-2b43db608d55",
		Name:           "Fix What's Broken",
		Completeness:   CompletenessFull,
		XMatters:       true,
		AdditionalCost: PayXLifeCost(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.X()
			return ReturnFromGraveyardTogether{
				Targets: graveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
					return (c.IsArtifact() || c.IsCreature()) && c.ManaValue() == x
				}),
				Controller: ctx.Controller(),
			}.Apply(ctx)
		},
	})
}
