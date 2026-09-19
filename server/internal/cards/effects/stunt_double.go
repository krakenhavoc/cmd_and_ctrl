package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stunt Double — Creature — Shapeshifter {3}{U}, 0/0:
//
//	"Flash
//	 You may have this creature enter as a copy of any creature on
//	 the battlefield."
//
// Clone with flash, and nothing else: the same candidate filter
// (anyCreatureOnBattlefield), no "except" clause. Flash rides
// PrintedKeywords so the cast-timing gate reads it from hand. It is
// NOT carried by the copy — a copy effect replaces the copiable
// values, and the catalog hooks, including the synthesized layer-6
// keyword static, key on the copied card's oracle ID afterwards
// (CR 707.2: the copy has exactly what the copied card prints).
//
// A declined or empty-board Stunt Double enters as its printed 0/0
// and dies (CR 704.5f), Clone's behaviour and for Clone's reason (see
// clone.go). It used to survive and declared the deviation; #691
// narrowed the state-based action's stand-in skip to objects with no
// printing behind them, and every clone got it back together.
func init() {
	Register(Spec{
		OracleID:        "a7ff1b64-8eab-4309-892c-17a619cd302e",
		Name:            "Stunt Double",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Stunt Double",
				anyCreatureOnBattlefield,
				nil,
			),
		},
	})
}
