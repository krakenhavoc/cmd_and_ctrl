package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mystic Zealot — Creature — Human Nomad Mystic {3}{W}, 2/4 (Odyssey,
// #5613):
//
//	"Threshold — As long as there are seven or more cards in your
//	 graveyard, this creature gets +1/+1 and has flying."
//
// Dreadwing Scavenger's threshold pair with flying for deathtouch: the
// layer 7c self-modifier (CR 613.4c) and a layer 6 self-grant, both read
// live on every recompute, so the flying comes and goes with the
// seventh card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5a47b99d-19dd-4830-89ee-35f4a74aa656",
		Name:         "Mystic Zealot",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{fraThresholdSelfPT(1, 1), fraThresholdSelfKeywords("flying")},
	})
}
