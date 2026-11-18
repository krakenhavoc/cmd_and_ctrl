package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	tap := AtYourEndStep("Cement Shoes — tap this creature", func(g *game.Game, item *game.StackItem) error {
		return TapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
	})
	Register(Spec{
		OracleID:     "52d04c36-9d38-42b2-9e47-905ab06cb708",
		Name:         "Cement Shoes",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       grantKey,
			Triggered: []game.TriggeredAbility{tap},
			Text:      "At the beginning of your end step, tap this creature.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(3, 3),
			GrantAbilitiesToAttached(grantKey),
		},
		UntapStepRestrictions: []game.UntapStepRestriction{
			{Label: "equipped creature doesn't untap", Restricts: AttachedToSource},
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
