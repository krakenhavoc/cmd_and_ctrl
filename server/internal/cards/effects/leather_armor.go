package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leather Armor — Artifact — Equipment {1} (Adventures in the Forgotten
// Realms, #5499):
//
//	"Equipped creature gets +0/+1 and has ward {1}.
//	 Equip {0}. Activate only once each turn."
//
// Armguard Familiar's equipped-creature ward (WardAttached) and Shredder's
// Armor's once-a-turn equip: the limit counts announcements of this
// Equipment's equip ability (CR 602.5b), so an equip that is countered
// still uses the turn's one.
//
// No simplification.
func init() {
	const equipLabel = "Equip {0}. Activate only once each turn."
	equip := EquipAbility("{0}")
	equip.Label = equipLabel
	equip.Condition = OncePerTurnActivation(equipLabel)
	Register(Spec{
		OracleID:     "7a6ce270-c261-422f-845f-11c22602a274",
		Name:         "Leather Armor",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(0, 1)},
		Triggered: []game.TriggeredAbility{
			WardAttached(WardMana("{1}"), "Leather Armor — equipped creature's ward {1}"),
		},
		Activated: []ActivatedAbility{equip},
	})
}
