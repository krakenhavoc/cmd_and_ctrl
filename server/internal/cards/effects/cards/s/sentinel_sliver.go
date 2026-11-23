package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sentinel Sliver — Creature — Sliver {1}{W}, 2/2 (EDHREC rank
// 4321):
//
//	"Sliver creatures you control have vigilance. (Attacking doesn't
//	 cause them to tap.)"
//
// Bonescythe Sliver's two-mana cousin, and in the batch alongside it
// for the reason two cards sharing a shape are always worth writing
// together: the shape composes or it doesn't, and three Slivers in
// this batch granting three different keywords off one builder is how
// that gets proved rather than asserted.
//
// No "other", so the Sentinel has vigilance itself; "you control", so
// the grant stops at your own board. Vigilance is a canonical engine
// keyword read by the attack declaration, so a Sliver that attacks
// under this stays untapped and can block on the crack back.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1d0f1186-be6c-45c3-9703-f0c1e13892fb",
		Name:         "Sentinel Sliver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, "vigilance"),
		},
	})
}
