package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Siege Smash — Instant {1}{R} (Modern Horizons 3, #5948):
//
//	"Split second (As long as this spell is on the stack, players can't
//	 cast spells or activate abilities that aren't mana abilities.)
//	 Choose one —
//	 • Destroy target artifact.
//	 • Target creature gets +3/+2 and gains trample until end of turn."
//
// Split second is the engine's keyword (CR 702.61); the modes are Abrade's
// artifact destruction and a pump with a keyword grant, each applying its
// own target clause.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4e07d30a-d29a-49bc-ab49-41ea4116a18c",
		Name:            "Siege Smash",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordSplitSecond},
		Purpose:         game.Purpose{Answers: game.AnswerRemove}, // the first mode destroys an artifact
		Modes: effects.ChooseOne(
			effects.Mode("Destroy target artifact.", effects.TargetPermanent("target artifact", effects.Artifact())),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets +3/+2 and gains trample until end of turn.",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return siegeSmashPump(ctx, t.ID)
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 3, 2, "trample"))),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if ctx.HasMode(0) && len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard {
				return effects.DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
			}
			return nil
		},
	})
}

func siegeSmashPump(ctx *effects.Context, id uuid.UUID) error {
	if err := (effects.BoostUntilEOT{Target: id, Power: 3, Toughness: 2, Label: "Siege Smash — +3/+2"}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Siege Smash — trample"}.Apply(ctx)
}
