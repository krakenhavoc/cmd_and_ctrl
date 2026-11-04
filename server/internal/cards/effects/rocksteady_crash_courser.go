package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rocksteady, Crash Courser — {4}{G}{G} Legendary Creature — Rhino Mutant 7/7
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"Rocksteady can't be blocked by more than one creature.
//	 Boars you control can't be blocked by more than one creature.
//	 Forestcycling {2} ({2}, Discard this card: Search your library for a
//	 Forest card, reveal it, put it into your hand, then shuffle.)"
//
// Both block limits are Outland Colossus's MaxBlockers, the second over the
// Boars its controller controls (Michelangelo, Mutant BFF's scope), read as
// blockers are declared (CR 509.1b). Forestcycling is typecycling (CR
// 702.29e) for any card with the Forest subtype.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f9330a7b-8a65-4e04-a777-bf6512332850",
		Name:         "Rocksteady, Crash Courser",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			MaxBlockers(OnSelf(), 1),
			MaxBlockers(OnMatching(And(Creature(), YouControl(), HasSubtype("Boar"))), 1),
		},
		Activated: []ActivatedAbility{
			Typecycling("Forestcycling", "{2}", "a Forest card", IsLandWithSubtype("Forest")),
		},
	})
}
