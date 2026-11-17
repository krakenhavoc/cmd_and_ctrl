package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mystic Enforcer — Creature — Human Nomad Mystic {2}{G}{W}, 3/3
// (Odyssey, #5613):
//
//	"Protection from black
//	 Threshold — As long as there are seven or more cards in your
//	 graveyard, this creature gets +3/+3 and has flying."
//
// Mystic Zealot's threshold pair at +3/+3 (the layer 7c self-modifier
// and the layer 6 self-grant, both read live), with a printed protection
// from black (CR 702.16).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cf3b79cc-ce97-48e9-a578-625e323d1b62",
		Name:            "Mystic Enforcer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.ProtectionFromColor("B")},
		Static:          []game.StaticAbility{fraThresholdSelfPT(3, 3), fraThresholdSelfKeywords("flying")},
	})
}
