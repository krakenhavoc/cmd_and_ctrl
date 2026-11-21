package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cement Shoes — Artifact — Equipment {1} (Streets of New Capenna, #5491):
//
//	"Equipped creature gets +3/+3 and has "At the beginning of your end
//	 step, tap this creature."
//	 Equipped creature doesn't untap during its controller's untap step.
//	 Equip {2}"
//
// The end-step trigger is a granted ability (Black Mage's Rod's shape): it
// belongs to the equipped creature, so "your" is the creature's controller
// and "this creature" is the creature. The untap lock is T-45 Power
// Armor's.
//
// No simplification.
func init() {
	const grantKey = "cement-shoes/end-step-tap"
	tap := effects.AtYourEndStep("Cement Shoes — tap this creature", func(g *game.Game, item *game.StackItem) error {
		return effects.TapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
	})
	effects.Register(effects.Spec{
		OracleID:     "52d04c36-9d38-42b2-9e47-905ab06cb708",
		Name:         "Cement Shoes",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       grantKey,
			Triggered: []game.TriggeredAbility{tap},
			Text:      "At the beginning of your end step, tap this creature.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 3),
			effects.GrantAbilitiesToAttached(grantKey),
		},
		UntapStepRestrictions: []game.UntapStepRestriction{
			{Label: "equipped creature doesn't untap", Restricts: effects.AttachedToSource},
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
