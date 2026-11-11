package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const armedWithProofPumpGrant = "armed-with-proof/clue-pump"
const armedWithProofEquipGrant = "armed-with-proof/clue-equip"

// Armed with Proof — Enchantment {2}{W} (Murders at Karlov Manor Commander):
//
//	"When this enchantment enters, investigate twice.
//	 Clues you control are Equipment in addition to their other types and
//	 have 'Equipped creature gets +2/+0' and equip {2}."
//
// Gemcutter Buccaneer's shape over Clues (granted_equipment.go): a layer-4
// Equipment subtype and a layer-6 grant of the pump static and the equip
// row. A Clue that is not a token counts (CR 205.3g).
//
// No simplification.
func init() {
	pump := PumpAttached(2, 0)
	pump.Label = "Equipped creature gets +2/+0."
	Register(Spec{
		OracleID:     "b401a490-d25f-4029-bb91-8999e140c6a1",
		Name:         "Armed with Proof",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{Key: armedWithProofPumpGrant, Static: []game.StaticAbility{pump}, Text: "Equipped creature gets +2/+0."},
			{Key: armedWithProofEquipGrant, Activated: []ActivatedAbility{EquipAbility("{2}")}, Text: "Equip {2}"},
		},
		Static: []game.StaticAbility{
			AreAlsoEquipment("Clues you control are Equipment in addition to their other types.", cmdrB1CluesYouControl),
			Labelled(GrantAbilities(cmdrB1CluesYouControl, armedWithProofPumpGrant, armedWithProofEquipGrant), "Clues you control have \"Equipped creature gets +2/+0\" and equip {2}."),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Armed with Proof — investigate twice", Do(Investigate{N: 2})),
		},
	})
}
