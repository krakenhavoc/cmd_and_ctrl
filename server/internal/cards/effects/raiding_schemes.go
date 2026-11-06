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
// Caveat: the creatures ride the cast's one teamwork_ids list, so a spell
// whose own cost already taps permanents there (New Blood, teamwork,
// Collective Effort's escalate, a replicate cost that taps) is not
// offered conspire. #4690.
func init() {
	Register(Spec{
		OracleID:     "854cac11-9a22-4dea-8c1b-c49a4858a754",
		Name:         "Raiding Schemes",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Spells whose own cost already taps permanents aren't given conspire.",
		},
		GrantedOptionalCosts: []game.GrantedOptionalCost{
			SpellsYouCastHave("Each noncreature spell you cast has conspire.", Noncreature(), Conspire()),
		},
	})
}
