package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dreamshackle Geist — Creature — Spirit {1}{U}{U}, 3/1 (Innistrad:
// Crimson Vow, S82):
//
//	"Flying
//	 At the beginning of combat on your turn, choose up to one —
//	 • Tap target creature.
//	 • Target creature doesn't untap during its controller's next untap
//	 step."
//
// Ferocification's modal beginning-of-combat trigger with Biblioplex
// Tomekeeper's "up to one" (ChooseN 0..1; CR 700.2). Choosing no mode
// removes the trigger as it would go on the stack (CR 700.2b). The second
// mode is DoesntUntapNextUntapStep, which names the target's controller's
// next untap step (CR 502.3); it does not tap the creature.
//
// No simplification.
func init() {
	t := effects.AtBeginningOfYourCombat("Dreamshackle Geist — choose up to one: tap target creature, or target creature doesn't untap during its controller's next untap step",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = effects.ChooseN("Choose up to one", 0, 1,
		effects.ModeDoing("Tap target creature.",
			effects.TargetCreature("target creature"),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				tg, ok := effects.ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return effects.TapTarget{Target: tg.ID}.Apply(ctx)
			}),
		effects.ModeDoing("Target creature doesn't untap during its controller's next untap step.",
			effects.TargetCreature("target creature"),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				tg, ok := effects.ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return effects.DoesntUntapNextUntapStep{Targets: []uuid.UUID{tg.ID}, Label: "Dreamshackle Geist — doesn't untap during its controller's next untap step"}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:        "aa40d270-63d7-4022-ade9-67c1ab65ef70",
		Name:            "Dreamshackle Geist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{t},
	})
}
