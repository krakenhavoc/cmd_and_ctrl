package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Retreat to Hagra — {2}{B} Enchantment:
//
//	"Landfall — Whenever a land you control enters, choose one —
//	 • Target creature gets +1/+0 and gains deathtouch until end of turn.
//	 • Each opponent loses 1 life and you gain 1 life."
//
// Felidar Retreat's shape: a landfall trigger whose own effect does nothing
// and whose two bullets are the bodies. The first bullet is Moment of
// Valor's until-end-of-turn record (+1/+0 in layer 7c, deathtouch in layer
// 6) on a target chosen as the trigger goes on the stack and re-checked as
// it resolves (CR 608.2b). The second names no target.
func init() {
	landfall := effects.Landfall("Retreat to Hagra — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	landfall.Modes = effects.ChooseOne(
		effects.ModeWithPurpose(effects.ModeDoing("Target creature gets +1/+0 and gains deathtouch until end of turn.",
			effects.TargetCreature("target creature"),
			effects.RfCreatureBModeTargetDoes(retreatToHagraPump)),
			effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "deathtouch"))),
		effects.ModeSteps("Each opponent loses 1 life and you gain 1 life.", nil,
			effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1}),
	)
	effects.Register(effects.Spec{
		OracleID:     "73cbb4f0-90d6-4228-9341-791e3e9c4b48",
		Name:         "Retreat to Hagra",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{landfall},
	})
}

// retreatToHagraPump is the first bullet on its still-legal target.
func retreatToHagraPump(ctx *effects.Context, id uuid.UUID) error {
	return effects.UntilEndOfTurn(ctx, id, nil, "Retreat to Hagra — +1/+0 and deathtouch",
		game.ModifyPTMod(1, 0), game.AddKeywordsMod("deathtouch"))
}
