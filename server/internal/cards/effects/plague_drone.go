package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Plague Drone — Creature — Demon {3}{B}, 3/3:
//
//	"Flying
//	 Rot Fly — If an opponent would gain life, that player loses that
//	 much life instead."
//
// Tainted Remedy on a flier. Rot Fly is a CR 614 replacement on the
// life an opponent would GAIN (LifeGainBecomes, opponents' scope),
// which since #482 every gain runs through — a catalog GainLife, a
// drain's gain half and lifelink alike. The delta is turned negative,
// so the opponent loses that much life and anything that watches life
// loss sees it. A loss can't be turned back into a gain by a second
// Drone: the window re-gathers, and a gain-only replacement no longer
// applies (CR 616.1). The Drone's controller's own gains are untouched.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9552fc97-9e6d-4307-bf58-30bf956356bf",
		Name:            "Plague Drone",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			LifeGainBecomes{
				Amount: func(n int) int { return -n },
				Scope:  LifeGainByOpponents,
				Label:  "Plague Drone: an opponent loses that much life instead",
			}.Build(),
		},
	})
}
