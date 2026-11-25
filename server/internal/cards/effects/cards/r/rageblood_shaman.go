package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rageblood Shaman — Creature — Minotaur Shaman {1}{R}{R}, 2/3 (Jumpstart,
// #5510):
//
//	"Trample
//	 Other Minotaur creatures you control get +1/+1 and have trample."
//
// Baru, Wurmspeaker's lord shape over Minotaurs: a layer-7c pump and a
// layer-6 keyword grant that read the effective subtypes, so a changeling
// counts, and neither touches the Shaman or an opponent's Minotaur.
//
// No simplification.
func init() {
	minotaurs := effects.TribeFilter{Tribes: []string{"Minotaur"}, Others: true, YoursOnly: true}
	effects.Register(effects.Spec{
		OracleID:        "2608f905-a678-4e40-9e73-cf1c42ec880d",
		Name:            "Rageblood Shaman",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			effects.TribalAnthem(minotaurs, 1, 1),
			effects.TribalKeywordGrant(minotaurs, "trample"),
		},
	})
}
