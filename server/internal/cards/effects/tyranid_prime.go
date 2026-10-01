package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tyranid Prime — Creature — Tyranid {1}{G}{U}, 0/4:
//
//	"Evolve (Whenever a creature you control enters, if that creature
//	 has greater power or toughness than this creature, put a +1/+1
//	 counter on this creature.)
//	 Synapse Creature — Other creatures you control have evolve."
//
// #1805's GRANT proof card (ADR 0106 §3). Evolve is a canonical keyword
// the engine turns into one trigger per instance (game/evolve.go), so
// the grant is an ordinary layer-6 keyword grant, and a granted evolve
// triggers exactly like a printed one. "Synapse Creature" is a flavor
// word with no rules meaning (CR 207.2d).
//
// Evolve is CUMULATIVE (CR 702.100d), and that is observable here: a
// creature that prints evolve has two instances under the Prime, and
// each triggers and re-checks on its own, so it can grow twice off one
// creature entering.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "12a43da3-57aa-4394-8cce-e53ec7e3538e",
		Name:            "Tyranid Prime",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordEvolve},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Others: true, YoursOnly: true}, game.KeywordEvolve),
		},
	})
}
