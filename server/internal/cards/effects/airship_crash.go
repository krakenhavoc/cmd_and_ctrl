package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Airship Crash — Instant {2}{G} (Final Fantasy, #4462):
//
//	"Destroy target artifact, enchantment, or creature with flying.
//	 Cycling {2}"
//
// "With flying" qualifies only the creature: any artifact or enchantment
// is a legal target, and a creature only while it has flying, judged at
// announce and again as the spell resolves (CR 608.2b). Cycling is the
// shared CR 702.29a ability from the hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "72a0e0f1-e8be-441a-bcaf-029fb1bd1fd2",
		Name:         "Airship Crash",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("target artifact, enchantment, or creature with flying",
			Or(Artifact(), Enchantment(), And(Creature(), HasKeyword("flying")))),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return destroyFirstLegalCardTarget(ctx.Game, item)
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
