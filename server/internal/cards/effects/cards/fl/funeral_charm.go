package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Funeral Charm — Instant {B} (Visions, S82 older sets batch 789, #5631):
//
//	"Choose one —
//	 • Target player discards a card.
//	 • Target creature gets +2/-1 until end of turn.
//	 • Target creature gains swampwalk until end of turn."
//
// Each bullet reads its own target group (CR 700.2): the player discards a
// card of their choice, the creature is pumped, or it gains swampwalk through
// the layer-6 grant the combat check reads (CR 702.14c). A target that is no
// longer legal does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8866272d-89cb-478c-82ae-20e1d510eab7",
		Name:         "Funeral Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Target player discards a card.", effects.TargetPlayer("target player"),
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer || t.ID == uuid.Nil {
						return nil
					}
					return effects.DiscardOne(ctx.Game, item, t.ID)
				}), effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 1})),
			effects.ModeDoing("Target creature gets +2/-1 until end of turn.", effects.TargetCreature("target creature"),
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: -1, Label: "Funeral Charm — +2/-1"}.Apply(ctx)
				}),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gains swampwalk until end of turn.", effects.TargetCreature("target creature"),
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"swampwalk"}, Label: "Funeral Charm — swampwalk"}.Apply(ctx)
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "swampwalk"))),
		),
	})
}
