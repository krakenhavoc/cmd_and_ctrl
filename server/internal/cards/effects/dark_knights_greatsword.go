package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dark Knight's Greatsword — Artifact — Equipment {2}{B} (Final
// Fantasy, #4462):
//
//	"Job select (When this Equipment enters, create a 1/1 colorless Hero
//	 creature token, then attach this to it.)
//	 Equipped creature gets +3/+0 and is a Knight in addition to its
//	 other types.
//	 Chaosbringer — Equip—Pay 3 life. Activate only once each turn."
//
// Dragoon's Lance's job select, pump and added Knight type, with an
// equip whose cost is life instead of mana (EquipPayingAbility, as
// Dissection Tools pays a sacrifice) limited to once each turn
// (OncePerTurnActivation: the announce tally is per object, so a second Greatsword has its own).
//
// No simplification.
func init() {
	const equipLabel = "Chaosbringer — Equip—Pay 3 life. Activate only once each turn."
	equip := EquipPayingAbility(equipLabel, PayLife(3))
	equip.Condition = OncePerTurnActivation(equipLabel)
	Register(Spec{
		OracleID:     "eda09ed4-7509-4d1e-b169-2b7c1c87e936",
		Name:         "Dark Knight's Greatsword",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(3, 0),
			AttachedIsAlso("Knight"),
		},
		Triggered: []game.TriggeredAbility{JobSelect("Dark Knight's Greatsword")},
		Activated: []ActivatedAbility{equip},
	})
}
