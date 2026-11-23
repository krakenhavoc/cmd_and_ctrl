package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guard Gomazoa — Creature — Jellyfish {2}{U}, 1/3 (Rise of the Eldrazi,
// #5581):
//
//	"Defender, flying
//	 Prevent all combat damage that would be dealt to this creature."
//
// Seraph of the Sword's prevention (CR 615.1a). Noncombat damage is dealt as
// usual.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7c565975-aebc-4599-ac32-5594c718e2cb",
		Name:            "Guard Gomazoa",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "flying"},
		Replacements:    []game.ReplacementEffect{effects.RoeB1PreventCombatDamageToThis("Guard Gomazoa")},
	})
}
