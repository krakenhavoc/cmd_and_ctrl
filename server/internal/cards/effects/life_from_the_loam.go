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
// (CR 608.2b) and the rest still return. Dredge is a draw replacement
// the catalog has no shape for yet (The Necrobloom's caveat says the
// same), so the card can be cast from hand but never dredges back.
func init() {
	Register(Spec{
		OracleID:     "ac8fba34-512c-4f24-999a-ab72f1ce4acb",
		Name:         "Life from the Loam",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Dredge isn't implemented — the card can't be returned from the graveyard in place of a draw."},
		Targets:      TargetCardInGraveyard("up to three target land cards from your graveyard", YouOwn(), Land()).WithCount(0, 3),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return returnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
