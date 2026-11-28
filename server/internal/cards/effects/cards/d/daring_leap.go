package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Daring Leap — Instant {1}{W}{U} (Planeshift, S82 OLD #5615):
//
//	"Target creature gets +1/+1 and gains flying and first strike until end of
//	 turn."
//
// One until-end-of-turn record carrying the +1/+1 (layer 7c) and both keyword
// grants (layer 6), on the target still legal as the spell resolves
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10bc00f7-d501-46e3-9f23-af1ddc3f6c4c",
		Name:         "Daring Leap",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "flying", "first strike")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return untilEOTGetsAndGains(ctx, id)
		},
	})
}

func untilEOTGetsAndGains(ctx *effects.Context, id uuid.UUID) error {
	return effects.UntilEndOfTurn(ctx, id, nil, "Daring Leap — +1/+1, flying and first strike",
		game.ModifyPTMod(1, 1), game.AddKeywordsMod("flying", "first strike"))
}
