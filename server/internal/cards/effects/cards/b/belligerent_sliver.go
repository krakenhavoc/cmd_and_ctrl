package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Belligerent Sliver — Creature — Sliver {2}{R}, 2/2 (Magic 2015, #5556):
//
//	"Sliver creatures you control have menace. (They can't be blocked
//	 except by two or more creatures.)"
//
// Cloudshredder Sliver's layer-6 grant over "Sliver creatures you
// control" (YoursOnly), the Belligerent Sliver itself included, read
// through effective subtypes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d54a535-e924-47c7-81db-e04df1c3f44e",
		Name:         "Belligerent Sliver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, "menace"),
		},
	})
}
