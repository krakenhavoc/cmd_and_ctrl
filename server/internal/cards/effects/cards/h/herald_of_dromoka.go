package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Herald of Dromoka — Creature — Human Warrior {1}{W}, 2/2 (Dragons of Tarkir):
//
//	"Vigilance
//	 Other Warrior creatures you control have vigilance."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dae8dc90-8047-434c-9e3a-a2eaa6aba575",
		Name:            "Herald of Dromoka",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Warrior"}, Others: true, YoursOnly: true}, "vigilance"),
		},
	})
}
