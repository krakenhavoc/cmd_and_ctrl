package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Decoy Ploy — {1}{B} Instant:
//
//	"Choose one or both —
//	 • Return target Villain card from your graveyard to your hand.
//	 • Return target Hero card from your graveyard to your hand."
//
// Two bullets, each with its own target group (ModeDoing); the spell is
// cast with whichever it chose and the engine skips a bullet whose target
// left (CR 608.2b).
//
// No simplification. Marvel Super Heroes (#4455).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "035cddbc-39c9-42af-9e26-c88419b43b49",
		Name:         "Decoy Ploy",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeDoing("Return target Villain card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target Villain card from your graveyard", effects.YouOwn(), effects.Subtype("Villain")),
				decoyPloyReturn),
			effects.ModeDoing("Return target Hero card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target Hero card from your graveyard", effects.YouOwn(), effects.Subtype("Hero")),
				decoyPloyReturn),
		),
	})
}

func decoyPloyReturn(_ *game.StackItem, ctx *effects.Context, occurrence int) error {
	target, ok := effects.ModeTarget(ctx, occurrence)
	if !ok || target.Kind != game.TargetCard {
		return nil
	}
	return effects.ReturnFromGraveyard{Target: target.ID, Dest: game.ZoneHand}.Apply(ctx)
}
