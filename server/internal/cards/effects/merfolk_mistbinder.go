package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merfolk Mistbinder — Creature — Merfolk Shaman {G}{U}, 2/2:
//
//	"Other Merfolk you control get +1/+1."
//
// "You control" keeps an opponent's Merfolk out; "other" keeps the
// Mistbinder itself out. Merfolk is read as an effective subtype, so a
// changeling is pumped.
func init() {
	Register(Spec{
		OracleID:     "be56dfea-1ece-4f06-b1cc-d7425b5018b4",
		Name:         "Merfolk Mistbinder",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(cmdr2B05MerfolkFilter(true, true), 1, 1),
		},
	})
}
