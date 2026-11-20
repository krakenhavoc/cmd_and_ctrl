package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wildwood Rebirth — Instant {1}{G} (Gatecrash, #5565):
//
//	"Return target creature card from your graveyard to your hand."
//
// Blood Beckoning's unkicked spell: the one creature card, if it is
// still in your graveyard as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7a4ddb8f-08ad-4bf9-82f6-72b073dd859a",
		Name:         "Wildwood Rebirth",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return returnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
