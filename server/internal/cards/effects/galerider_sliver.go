package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Galerider Sliver — 1/1 Creature — Sliver for {U} (EDHREC rank
// 4042):
//
//	"Sliver creatures you control have flying."
//
// The one-mana evasion Sliver. It is in the batch as the keyword half
// of the lord builder — TribalKeywordGrant rather than TribalAnthem —
// and as the Sliver-specific wrinkle worth checking: modern Slivers
// say "you control" where the originals (Lord of Atlantis, Winged
// Sliver) did not, so this one must NOT hand flying to the Sliver deck
// across the table.
//
// It grants to itself as well — there is no "other" in the printed
// text — so a lone Galerider Sliver flies.
//
// Layer 6, over post-layer types, so a changeling counts as a Sliver
// and gets flying too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1ad78038-44f9-4599-84db-fc1a87d9ed45",
		Name:         "Galerider Sliver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, "flying"),
		},
	})
}
