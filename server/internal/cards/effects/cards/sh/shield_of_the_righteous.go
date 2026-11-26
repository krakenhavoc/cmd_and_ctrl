package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shield of the Righteous — Artifact — Equipment {W}{U} (Alara Reborn, #5586):
//
//	"Equipped creature gets +0/+2 and has vigilance.
//	 Whenever equipped creature blocks a creature, that creature doesn't untap
//	 during its controller's next untap step.
//	 Equip {2}"
//
// Accorder's Shield's two statics, and Labyrinth Minotaur's block trigger for the
// creature the Equipment is attached to (one trigger per creature it blocks,
// CR 509.3b); the attacker is recorded as the trigger goes on the stack with its
// CR 400.7 epoch.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6d9aaa70-b2cc-424c-8095-a0b2fa272db3",
		Name:         "Shield of the Righteous",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(0, 2),
			effects.GrantToAttached("vigilance"),
		},
		Triggered: []game.TriggeredAbility{
			effects.S82OldB504EquippedBlocksItDoesntUntap("Shield of the Righteous — that creature doesn't untap during its controller's next untap step"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
