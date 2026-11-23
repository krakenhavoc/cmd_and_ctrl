package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Adaptive Sporesinger — Creature — Phyrexian Druid {2}{G}, 2/2 (Phyrexia:
// All Will Be One, S82):
//
//	"Vigilance
//	 When this creature enters, choose one —
//	 • Target creature gets +2/+2 and gains vigilance until end of turn.
//	 • Proliferate."
//
// A modal enters trigger (CR 603.3c): the mode is chosen as the trigger goes
// on the stack, and only the first bullet targets. A target gone by
// resolution leaves the pump with nothing to do (CR 608.2b).
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Adaptive Sporesinger — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = effects.ChooseOne(
		effects.ModeDoing("Target creature gets +2/+2 and gains vigilance until end of turn.",
			effects.TargetCreature("target creature"),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Adaptive Sporesinger — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"vigilance"},
					Label:    "Adaptive Sporesinger — vigilance until end of turn",
				}.Apply(ctx)
			}),
		effects.ModeSteps("Proliferate.", nil, effects.Proliferate{}),
	)
	effects.Register(effects.Spec{
		OracleID:        "f6311b7a-7482-4b13-9f8f-13e80c98e98e",
		Name:            "Adaptive Sporesinger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
