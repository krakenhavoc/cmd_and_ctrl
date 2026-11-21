package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const supernaturalStaminaGrant = "supernatural-stamina/return-tapped"

// Supernatural Stamina — Instant {B} (Commander Masters, #5948):
//
//	"Until end of turn, target creature gets +2/+0 and gains 'When this
//	 creature dies, return it to the battlefield tapped under its owner's
//	 control.'"
//
// The creature gets the pump and the dies trigger together, both until end
// of turn. The trigger returns the very object that died (CR 400.7), under
// its owner's control and tapped, and does nothing for a creature that has
// already left the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "65756e6d-9825-4ed6-91c4-79f1e3bd4ae5",
		Name:         "Supernatural Stamina",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0)),
		Grants: []effects.AbilityGrant{{
			Key: supernaturalStaminaGrant,
			Triggered: []game.TriggeredAbility{
				effects.WhenThisDies("Supernatural Stamina — return it to the battlefield tapped under its owner's control",
					effects.ReturnThisCreatureFromGraveyard(nil, nil)),
			},
			Text: "When this creature dies, return it to the battlefield tapped under its owner's control.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: target, Power: 2, Label: "Supernatural Stamina — +2/+0"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantAbilitiesFor{Target: target, Keys: []string{supernaturalStaminaGrant},
				Label: "Supernatural Stamina — until end of turn, it returns when it dies"}.Apply(ctx)
		},
	})
}
