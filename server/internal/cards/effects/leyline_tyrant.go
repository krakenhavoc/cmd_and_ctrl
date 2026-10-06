package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leyline Tyrant — Creature — Dragon {2}{R}{R}, 4/4:
//
//	"Flying
//	 You don't lose unspent red mana as steps and phases end.
//	 When this creature dies, you may pay any amount of {R}. When you
//	 do, it deals that much damage to any target."
//
// The mana clause is Spec.ManaPool (#2166): red only, derived from the
// battlefield, so it ends the moment the Tyrant leaves. The dies trigger
// is not built: it needs a variable payment of {R} from a creature that
// has already died, then a reflexive trigger. The card is weaker than
// printed without it, never stronger.
func init() {
	Register(Spec{
		OracleID:     "f92aaa00-6ece-4033-b3df-b2fc2c4718d9",
		Name:         "Leyline Tyrant",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The dies trigger that lets you pay any amount of {R} to deal that much damage isn't implemented — only the flying and keeping red mana work.",
		},
		PrintedKeywords: []string{"flying"},
		ManaPool: []game.ManaPoolStatic{
			{Kind: game.ManaPoolKeep, Colors: []string{"R"}},
		},
	})
}
