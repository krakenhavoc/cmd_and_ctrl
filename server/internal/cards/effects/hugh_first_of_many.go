package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hugh, First of Many — Legendary Artifact Creature — Borg {1}{W}{U}{B},
// 3/4 (Star Trek Commander, #4974):
//
//	"Artifact spells you cast have replicate {2}{U}. (Whenever you cast an
//	 artifact spell, copy it for each time you paid its replicate cost.
//	 Each copy becomes a token.)"
//
// Replicate is CR 702.56a's two halves. The cost is a granted optional
// cost (granted_optional_cost.go, Silverquill's casualty shape): each
// artifact spell Hugh's controller casts while Hugh is on the battlefield
// may pay {2}{U} any number of times. The engine caps one announcement at
// 20 payments, multikicker's reason; mana bounds it far below that. The
// trigger is the one the spell gains, "when you cast this spell, if a
// replicate cost was paid for it, copy it for each time", built on Hugh
// because the spell had the ability as it was cast; the count is read
// then (CR 603.4) and the spell is named, not targeted, so one countered
// in response is still copied (CR 608.2h). Each copy of the artifact spell
// becomes a token as it resolves (CR 111.13, CR 608.3f). With two Hughs
// the spell is offered one replicate cost, and only one copies it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9964f2fe-b7f4-40f5-8e71-948756d27f0a",
		Name:         "Hugh, First of Many",
		Completeness: CompletenessFull,
		GrantedOptionalCosts: []game.GrantedOptionalCost{SpellsYouCastHave(
			"Artifact spells you cast have replicate {2}{U}.", Artifact(), replicateMana("{2}{U}", 20))},
		Triggered: []game.TriggeredAbility{
			trekB39HughReplicate("Hugh, First of Many — replicate: copy the artifact spell"),
		},
	})
}
