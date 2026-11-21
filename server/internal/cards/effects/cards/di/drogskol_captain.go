package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drogskol Captain — Creature — Spirit Soldier {1}{W}{U}, 2/2:
//
//	"Flying
//	 Other Spirit creatures you control get +1/+1 and have hexproof."
//
// A lord scoped to the OTHER Spirits its controller controls: the shared
// tribal anthem (layer 7c) and keyword grant (layer 6), so a Spirit token
// made later is a 2/2 hexproof the moment it exists and the Captain itself
// stays a plain 2/2 with no hexproof.
//
// No simplification.
func init() {
	spirits := effects.TribeFilter{Tribes: []string{"Spirit"}, Others: true, YoursOnly: true}
	effects.Register(effects.Spec{
		OracleID:        "c33604ac-636a-43ad-9efe-8d7b3a28bf81",
		Name:            "Drogskol Captain",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			effects.TribalAnthem(spirits, 1, 1),
			effects.TribalKeywordGrant(spirits, "hexproof"),
		},
	})
}
