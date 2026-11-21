package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "65756e6d-9825-4ed6-91c4-79f1e3bd4ae5",
		Name:         "Supernatural Stamina",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 0)),
		Grants: []AbilityGrant{{
			Key: supernaturalStaminaGrant,
			Triggered: []game.TriggeredAbility{
				WhenThisDies("Supernatural Stamina — return it to the battlefield tapped under its owner's control",
					returnThisCreatureFromGraveyard(nil, nil)),
			},
			Text: "When this creature dies, return it to the battlefield tapped under its owner's control.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (BoostUntilEOT{Target: target, Power: 2, Label: "Supernatural Stamina — +2/+0"}).Apply(ctx); err != nil {
				return err
			}
			return GrantAbilitiesFor{Target: target, Keys: []string{supernaturalStaminaGrant},
				Label: "Supernatural Stamina — until end of turn, it returns when it dies"}.Apply(ctx)
		},
	})
}
