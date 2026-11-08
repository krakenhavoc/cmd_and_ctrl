package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dauntless Dismantler — Creature — Human Artificer {1}{W}, 1/4 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Artifacts your opponents control enter tapped.
//	 {X}{X}{W}, Sacrifice this creature: Destroy each artifact with mana
//	 value X."
//
// The static is Manglehorn's CR 614 replacement
// (opponentsArtifactsEnterTapped): an artifact entering under another
// player's control, token or not, arrives tapped.
//
// "{X}{X}" is two X slots, Blast Zone's, so X=2 costs {4}{W}. The
// sacrifice is a cost, so the Dismantler is gone before the ability
// resolves. The sweep reads each artifact's mana value on the
// battlefield, where an {X} in its cost counts as 0 (CR 202.3e) and a
// token's is 0, so X=0 destroys Treasures and other tokens. It is one
// simultaneous destroy (DestroyAllMatching), it hits your own artifacts
// too, and an indestructible artifact survives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7129bdec-83e0-4c1e-a87b-ac0da1a714b3",
		Name:         "Dauntless Dismantler",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			opponentsArtifactsEnterTapped("Dauntless Dismantler: artifacts your opponents control enter tapped"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{X}{X}{W}, Sacrifice this creature: Destroy each artifact with mana value X.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepArtifacts, How: game.SweepDestroy, Partial: true}},
			Cost:    Plus(ManaCost("{X}{X}{W}"), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := ctx.X()
				return DestroyAllMatching{Match: And(Artifact(), ManaValueGE(x), ManaValueLE(x))}.Apply(ctx)
			},
		}},
	})
}
