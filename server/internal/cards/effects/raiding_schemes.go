package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raiding Schemes — Enchantment {3}{R}{G} (Lorwyn Eclipsed):
//
//	"Each noncreature spell you cast has conspire. (As you cast a
//	 noncreature spell, you may tap two untapped creatures you control
//	 that share a color with it. When you do, copy it and you may choose
//	 new targets for the copy. A copy of a permanent spell becomes a
//	 token.)"
//
// The grant is a GrantedOptionalCost (game/granted_optional_cost.go):
// while Raiding Schemes is on the battlefield under your control, every
// noncreature spell you cast offers "Conspire" beside its own optional
// costs. The two creatures are tapped as the spell is cast, each sharing
// a colour with it (CR 702.78a), and the copy is the spell's own conspire
// trigger (game/conspire.go), read off what was paid. A copy of an
// artifact, enchantment or planeswalker spell becomes a token as it
// resolves (CR 707.10f).
//
// The two creatures ride the cast's conspire_ids, a list of their own
// (#5688), so a spell whose own cost taps permanents (New Blood, teamwork,
// Collective Effort's escalate, a replicate cost that taps) is conspired
// by tapping two more: one creature never pays both. Two Raiding Schemes
// give a spell two instances, each paid and copied on its own (CR
// 702.78b).
func init() {
	Register(Spec{
		OracleID:     "854cac11-9a22-4dea-8c1b-c49a4858a754",
		Name:         "Raiding Schemes",
		Completeness: CompletenessFull,
		GrantedOptionalCosts: []game.GrantedOptionalCost{
			ConspireForSpellsYouCast("Each noncreature spell you cast has conspire.", Noncreature()),
		},
	})
}
