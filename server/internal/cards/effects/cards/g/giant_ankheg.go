package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Giant Ankheg — Creature — Insect {6}{G}{G}, 8/8 (Commander Legends:
// Battle for Baldur's Gate):
//
//	"Trample
//	 Ward {2}
//	 Other creatures you control have trample and ward {2}."
//
// Trample rides PrintedKeywords; the printed ward is Ward and the granted
// ward is WardGranted over the other creatures its controller controls
// (the Hexing Squelcher shape), with trample granted by KeywordGrant
// (CR 702.19, CR 702.21).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ca905672-50c2-4e19-9d08-2d3f8e3debfd",
		Name:            "Giant Ankheg",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.OtherCreaturesYouControl, "trample"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Giant Ankheg — ward {2}"),
			effects.WardGranted(effects.WardMana("{2}"), "Giant Ankheg — other creatures you control have ward {2}",
				effects.TribeFilter{Others: true, YoursOnly: true}.Matches),
		},
	})
}
