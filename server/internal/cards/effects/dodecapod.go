package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dodecapod — Artifact Creature — Golem {4}, 3/3:
//
//	"If a spell or ability an opponent controls causes you to discard
//	 this card, put it onto the battlefield with two +1/+1 counters on
//	 it instead of putting it into your graveyard."
//
// The Baloth replacement (ADR 0153, see obstinate_baloth.go) with
// "with two +1/+1 counters on it" as a CR 614.1c enters-with clause: the
// counters ride the entry event, so a Hardened Scales-style modifier
// sees them and the permanent is a 5/5 the moment it is on the
// battlefield, never a 3/3 that grows afterwards.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "2c189e3b-90d3-49b4-bebe-a4f14c8a275b",
		Name:              "Dodecapod",
		Completeness:      CompletenessFull,
		DiscardByOpponent: game.DiscardByOpponentBattlefield,
		Replacements: []game.ReplacementEffect{
			OntoTheBattlefieldWhenAnOpponentMakesYouDiscardThis("Dodecapod",
				map[string]int{game.CounterPlusOne: 2}),
		},
	})
}
