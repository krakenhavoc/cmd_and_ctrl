package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Auroral Procession — Instant {G}{U} (EDHREC rank 3583):
//
//	"Return target card from your graveyard to your hand."
//
// Regrowth at instant speed, any card type. A target that left the
// graveyard in response fizzles the spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "511d3de9-076b-4811-bddf-00623f98993f",
		Name:         "Auroral Procession",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target card from your graveyard", YouOwn()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b34ReturnChosenGraveyardCardToHand(ctx)
		},
	})
}
