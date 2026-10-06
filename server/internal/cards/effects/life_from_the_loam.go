package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Life from the Loam — Sorcery {1}{G}:
//
//	"Return up to three target land cards from your graveyard to your
//	 hand.
//	 Dredge 3 (If you would draw a card, you may mill three cards
//	 instead. If you do, return this card from your graveyard to your
//	 hand.)"
//
// The spell half is Blood Beckoning's return with a count of "up to
// three"; a target that left the graveyard in response is skipped
// (CR 608.2b) and the rest still return. Dredge 3 is dredge.go (#2127):
// from the graveyard, with three or more cards in the library, the card
// is offered in place of a draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ac8fba34-512c-4f24-999a-ab72f1ce4acb",
		Name:         "Life from the Loam",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{Dredge(3)},
		Targets:      TargetCardInGraveyard("up to three target land cards from your graveyard", YouOwn(), Land()).WithCount(0, 3),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return returnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
