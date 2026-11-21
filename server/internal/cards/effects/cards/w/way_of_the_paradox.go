package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Way of the Paradox — Legendary Enchantment {2}{G} (Reality Fracture,
// tracker #2795):
//
//	"When Way of the Paradox enters, empower Jace 5.
//	 Whenever you activate a loyalty ability, you gain 1 life. You may
//	 play an additional land this turn."
//
// Empower Jace is the keyword action (ADR 0139). The trigger is one per
// loyalty activation, printed or granted; the extra land drop stacks
// with the turn's others, as Explore's does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76771cd1-6b74-49bb-9421-8845fe344567",
		Name:         "Way of the Paradox",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{ExtraLandDrops: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Way of the Paradox — empower Jace 5", effects.Do(effects.EmpowerJace{N: 5})),
			effects.On(game.EventActivateAbility, effects.FrYouActivatedALoyaltyAbility,
				"Way of the Paradox — gain 1 life and play an additional land",
				// The land drop waits for the life gain's prompt (a CR 616 order; CR 608.2c, #4155).
				effects.Do(
					effects.GainLife{Amount: 1},
					effects.Instruction(func(ctx *effects.Context) error {
						ctx.Game.GrantAdditionalLandPlayForEffect(ctx.Controller(), 1)
						return nil
					}),
				)),
		},
	})
}
