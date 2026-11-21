package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oathsworn Giant — Creature — Giant Soldier {4}{W}{W}, 3/4 (Ravnica: City
// of Guilds, #5600):
//
//	"Vigilance
//	 Other creatures you control get +0/+2 and have vigilance."
//
// Vigilance is printed data; the anthem is layer 7c and the keyword grant
// layer 6, both over the other creatures its controller controls.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d3828907-1321-40ca-9531-30401e82a6b8",
		Name:            "Oathsworn Giant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Others: true, YoursOnly: true}, 0, 2),
			effects.TribalKeywordGrant(effects.TribeFilter{Others: true, YoursOnly: true}, "vigilance"),
		},
	})
}
