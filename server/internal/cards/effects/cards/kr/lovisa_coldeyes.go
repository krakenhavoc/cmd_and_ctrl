package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lovisa Coldeyes — Legendary Creature — Human {3}{R}{R}, 3/3 (Coldsnap, S82 OLD b560):
//
//	"Each creature that's a Barbarian, a Warrior, or a Berserker gets +2/+2 and
//	 has haste."
//
// A tribal lord over three types, whoever controls the creature, read through
// the layered view; Lovisa herself is a Human and gets nothing.
//
// No simplification.
func init() {
	tribes := effects.TribeFilter{Tribes: []string{"Barbarian", "Warrior", "Berserker"}}
	effects.Register(effects.Spec{
		OracleID:     "720879dc-7593-41c0-ab3d-936069f26930",
		Name:         "Lovisa Coldeyes",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(tribes, 2, 2),
			effects.TribalKeywordGrant(tribes, "haste"),
		},
	})
}
