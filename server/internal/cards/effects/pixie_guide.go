package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pixie Guide — Creature — Faerie 1/3 {1}{U} (Adventures in the Forgotten
// Realms, #5681):
//
//	"Flying
//	 Grant an Advantage — If you would roll one or more dice, instead roll
//	 that many dice plus one and ignore the lowest roll."
//
// The ability is a replacement on the dice roll's window (CR 614.1a,
// RollOneMoreDieAndIgnoreTheLowest): every roll its controller makes
// rolls one more die and ignores the lowest. An ignored roll never
// happened (CR 706.6): it triggers nothing and no results table reads it.
// Two Pixie Guides roll two more dice and ignore the two lowest. The
// planar die has no numerical result, so it has no lowest roll and is
// left alone (CR 706.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b6003419-1a2a-46f3-b32c-47b825324f78",
		Name:         "Pixie Guide",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{RollOneMoreDieAndIgnoreTheLowest("Pixie Guide")},
	})
}
