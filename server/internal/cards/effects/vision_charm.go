package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vision Charm — Instant {U}:
//
//	"Choose one —
//	 • Target player mills four cards.
//	 • Choose a land type and a basic land type. Each land of the first
//	   chosen type becomes the second chosen type until end of turn.
//	 • Target artifact phases out. (While it's phased out, it's treated as
//	   though it doesn't exist. It phases in before its controller untaps
//	   during their next untap step.)"
//
// The second bullet is ADR 0109 §1 decision 4 (#1881): two choices as it
// resolves (CR 608.2), the first over all seventeen land types of CR
// 205.3i (Gate, Desert and Urza's included) and the second over the five
// basic ones. The lands are the ones that have the first type then (CR
// 611.2c). Until end of turn each one's land types are replaced by the
// second (other subtypes stay, CR 205.1a), it loses the abilities its
// rules text gives it and taps for the new colour (CR 305.6). Choosing
// the same basic type twice is legal, and still strips the rules text of
// every land of that type (a Breeding Pool named as a Forest becomes a
// plain Forest).
//
// The third bullet is the phasing primitive (CR 702.26).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c56d2c5b-c3e3-4236-bc38-286e250aa4f7",
		Name:         "Vision Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Target player mills four cards.", TargetPlayer("target player"),
				visionCharmMill),
			ModeDoing("Choose a land type and a basic land type. Each land of the first chosen type becomes the second chosen type until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return ChooseBasicLandTypeThen(ctx, "Vision Charm — choose a land type", game.LandTypes, func(ctx *Context, from string) error {
						return ChooseBasicLandTypeThen(ctx, "Vision Charm — choose a basic land type", nil, func(ctx *Context, to string) error {
							return LandBecomes{
								Match:    And(Land(), HasSubtype(from)),
								Types:    []string{to},
								Duration: DurationUntilEndOfTurn(ctx),
								Label:    "Vision Charm — each " + from + " is " + landTypeArticle(to),
							}.Apply(ctx)
						})
					})
				}),
			ModeDoing("Target artifact phases out. (While it's phased out, it's treated as though it doesn't exist. It phases in before its controller untaps during their next untap step.)",
				TargetPermanent("target artifact", Artifact()),
				visionCharmPhaseOut),
		),
	})
}

// visionCharmMill is the first bullet: the target player mills four.
func visionCharmMill(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	return MillCards{Player: t.ID, N: 4}.Apply(ctx)
}

// visionCharmPhaseOut is the third bullet: the target artifact phases
// out.
func visionCharmPhaseOut(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	return PhaseOut{Targets: []uuid.UUID{t.ID}}.Apply(ctx)
}
