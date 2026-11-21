package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	equip := effects.EquipAbility("{0}")
	equip.Label = equipLabel
	equip.Condition = effects.OncePerTurnActivation(equipLabel)
	effects.Register(effects.Spec{
		OracleID:     "7a6ce270-c261-422f-845f-11c22602a274",
		Name:         "Leather Armor",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(0, 1)},
		Triggered: []game.TriggeredAbility{
			effects.WardAttached(effects.WardMana("{1}"), "Leather Armor — equipped creature's ward {1}"),
		},
		Activated: []effects.ActivatedAbility{equip},
	})
}
