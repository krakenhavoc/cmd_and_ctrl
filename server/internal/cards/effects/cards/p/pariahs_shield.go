package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pariah's Shield — Artifact — Equipment {5}:
//
//	"All damage that would be dealt to you is dealt to equipped creature
//	 instead.
//	 Equip {3}"
//
// ADR 0108 §9 decision 4 (#1905): Pariah's static redirection on an
// Equipment. The ruling: unattached, damage is dealt to you as normal.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f2103ab8-a183-4db8-98dd-4146217b5125",
		Name:         "Pariah's Shield",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.RedirectYourDamageToAttached("Pariah's Shield — damage to you is dealt to equipped creature instead"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
