package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Adventuring Gear — Artifact — Equipment {1} (Foundations, #4465):
//
//	"Landfall — Whenever a land you control enters, equipped creature
//	 gets +2/+2 until end of turn.
//	 Equip {1}"
//
// Landfall is Rampaging Baloths' trigger. "Equipped creature" is read as
// the trigger resolves, Umezawa's Jitte's pump: the creature the Gear is
// attached to then, from the Gear's last-known information if it has left
// (CR 608.2h). Unattached, the trigger still goes on the stack and pumps
// nothing. Two land drops in a turn are +4/+4.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "17042684-7af4-43a6-88c7-885032cfb27c",
		Name:         "Adventuring Gear",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Adventuring Gear — equipped creature gets +2/+2 until end of turn",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.PumpEquippedCreatureUntilEOT(ctx, 2, 2, "Adventuring Gear — +2/+2")
				}))),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
