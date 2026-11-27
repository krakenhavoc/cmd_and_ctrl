package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Living Airship — Creature — Metathran {3}{U}, 2/3 (Apocalypse, S82 OLD,
// #5614):
//
//	"Flying
//	 {2}{G}: Regenerate this creature."
//
// A printed keyword and the shared regenerate row (CR 701.19a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3bcead90-f7be-4fba-84e7-3d7c788a8b64",
		Name:            "Living Airship",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.RegenerateThisRow("{2}{G}")},
	})
}
