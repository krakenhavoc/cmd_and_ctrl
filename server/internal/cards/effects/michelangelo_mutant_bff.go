package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Michelangelo, Mutant BFF — {2}{G}{G} Legendary Creature — Mutant Ninja
// Turtle 4/4 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Each creature you control with a counter on it can't be blocked by
//	 more than one creature.
//	 Whenever Michelangelo enters or attacks, create a Mutagen token.
//	 (It's an artifact with "{1}, {T}, Sacrifice this token: Put a +1/+1 counter on target
//	 creature. Activate only as a sorcery.")"
//
// The block limit is Outland Colossus's MaxBlockers, over every creature
// its controller controls with a counter of any kind on it (CR 122.1),
// read as blockers are declared (CR 509.1b), so a creature that gets its
// first counter after blocks is not affected by this combat. The trigger
// is one ability with two conditions (WhenThisEntersOrAttacks). The Mutagen token is the catalog's CR 111.10v template (MutagenToken, #4676).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fe1eb8b3-c605-4711-a057-ed3565bfe75c",
		Name:         "Michelangelo, Mutant BFF",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			MaxBlockers(OnMatching(And(Creature(), YouControl(), b41HasACounter())), 1),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Michelangelo, Mutant BFF — create a Mutagen token",
				Do(CreateToken{Template: MutagenToken(), N: 1})),
		},
	})
}
