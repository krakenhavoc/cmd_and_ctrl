package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "035cddbc-39c9-42af-9e26-c88419b43b49",
		Name:         "Decoy Ploy",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose one or both", 1, 2,
			ModeDoing("Return target Villain card from your graveyard to your hand.",
				TargetCardInGraveyard("target Villain card from your graveyard", YouOwn(), Subtype("Villain")),
				decoyPloyReturn),
			ModeDoing("Return target Hero card from your graveyard to your hand.",
				TargetCardInGraveyard("target Hero card from your graveyard", YouOwn(), Subtype("Hero")),
				decoyPloyReturn),
		),
	})
}

func decoyPloyReturn(_ *game.StackItem, ctx *Context, occurrence int) error {
	target, ok := ModeTarget(ctx, occurrence)
	if !ok || target.Kind != game.TargetCard {
		return nil
	}
	return ReturnFromGraveyard{Target: target.ID, Dest: game.ZoneHand}.Apply(ctx)
}
