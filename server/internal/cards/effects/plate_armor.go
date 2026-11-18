package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Plate Armor — Artifact — Equipment {2}{W} (Adventures in the Forgotten
// Realms, #5499):
//
//	"Equipped creature gets +3/+3 and has ward {1}.
//	 Equip {3}. This ability costs {1} less to activate for each other
//	 Equipment you control."
//
// Bladehold War-Whip's equip reduction, scaled by the other Equipment its
// controller has (ActivationCostsLessEach counts every battlefield card but
// the one being priced) and limited to this Equipment's own abilities
// (thisPermanentsAbility). Only the generic part can be reduced
// (CR 601.2f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "36082b5e-760b-4245-af59-00d308470d1a",
		Name:         "Plate Armor",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(3, 3)},
		Triggered: []game.TriggeredAbility{
			WardAttached(WardMana("{1}"), "Plate Armor — equipped creature's ward {1}"),
		},
		CostModifiers: []game.CostModifier{
			ActivationCostsLessEach(PermanentsYouControl(OfSubtype("Equipment")),
				"Equip {3}. This ability costs {1} less to activate for each other Equipment you control.",
				AnEquipAbilityCost(), thisPermanentsAbility(), ActivatedByTheModifiersController()),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
