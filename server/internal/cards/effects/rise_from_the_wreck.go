package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rise from the Wreck — {2}{G} Sorcery (Aetherdrift, #4464):
//
//	"Return up to one target creature card, up to one target Mount card,
//	 up to one target Vehicle card, and up to one target creature card
//	 with no abilities from your graveyard to your hand."
//
// Four clauses, each "up to one" (CR 601.2c), all from your graveyard. The
// same card may fill two of them, since each "target" is its own instance
// of the word (CR 115.3): a vanilla creature is both a creature card and a
// creature card with no abilities, and returns once. "No abilities" reads
// the printed text (game.PrintsNoAbilities).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8eeb83b4-dca1-497e-a6b7-0485807aeef3",
		Name:         "Rise from the Wreck",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCardInGraveyard("up to one target creature card from your graveyard", YouOwn(), Creature()).WithCount(0, 1),
			TargetCardInGraveyard("up to one target Mount card from your graveyard", YouOwn(), OfSubtype("Mount")).WithCount(0, 1),
			TargetCardInGraveyard("up to one target Vehicle card from your graveyard", YouOwn(), OfSubtype("Vehicle")).WithCount(0, 1),
			TargetCardInGraveyard("up to one target creature card with no abilities from your graveyard", YouOwn(), b13NoAbilities()).WithCount(0, 1),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return returnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
