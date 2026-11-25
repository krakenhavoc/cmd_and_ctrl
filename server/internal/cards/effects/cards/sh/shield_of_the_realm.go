package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shield of the Realm — Artifact — Equipment {2} (Dominaria, S82 OLD batch
// 214, #5530):
//
//	"If a source would deal damage to equipped creature, prevent 2 of that
//	 damage.
//	 Equip {1}"
//
// A prevention that takes 2 off each instance (CR 615): damage of 2 or less is
// prevented entirely, larger damage loses 2. It protects whatever the Shield is
// attached to as the damage would be dealt. Several sources are separate
// instances, so each loses 2.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d669ed30-efe9-4455-813a-c747b123333f",
		Name:         "Shield of the Realm",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.OldB214PreventOfDamageToEquipped(2, "Shield of the Realm: prevent 2 of that damage"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
