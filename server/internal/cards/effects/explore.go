package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Explore — Sorcery {1}{G}:
//
//	"You may play an additional land this turn.
//	 Draw a card."
//
// The two-mana cantrip that turns a land in hand into ramp. Its
// blocker on the batch-01 triage was "player / game-rule statics":
// the land-drop limit was not enforced at all, so an extra drop had
// nothing to add to. #500 made the limit real and gave it two
// suppliers — a permanent's standing "on each of your turns"
// (Spec.AdditionalLandPlays, Exploration and Azusa) and a one-shot
// grant for the rest of THIS turn, which is this card.
//
// `GrantAdditionalLandPlayForEffect` is the one-shot half. The grant
// expires with the rest of the per-turn state at the next turn
// boundary, so an Explore cast on an opponent's turn gives a drop
// that seat cannot use — printed behaviour, and the reason the card
// is a sorcery.
//
// Two Explores stack, because the grant adds rather than sets.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b8f566ea-8283-4afa-9ac8-737e26419283",
		Name:         "Explore",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ctx.Game.GrantAdditionalLandPlayForEffect(item.Controller, 1)
			return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
