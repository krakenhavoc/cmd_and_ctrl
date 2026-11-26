package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shield of the Avatar — Artifact — Equipment {1} (Magic 2015, #5556):
//
//	"If a source would deal damage to equipped creature, prevent X of that
//	 damage, where X is the number of creatures you control.
//	 Equip {2}"
//
// Shield of the Realm's prevention (CR 615) with a number that is read each
// time damage is about to be dealt: the creatures its controller controls
// then, the equipped creature among them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a4433da6-78b7-4669-a0dd-e7ea9672b0bc",
		Name:         "Shield of the Avatar",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.S82OldB357PreventXDamageToEquipped("Shield of the Avatar: prevent X of that damage"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
