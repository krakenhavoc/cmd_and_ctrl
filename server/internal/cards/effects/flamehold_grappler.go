package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flamehold Grappler — Creature — Human Monk {U}{R}{W}, 3/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"First strike
//	 When this creature enters, copy the next spell you cast this turn when
//	 you cast it. You may choose new targets for the copy. (A copy of a
//	 permanent spell becomes a token.)"
//
// The enters trigger creates a delayed triggered ability (CR 603.7a), the
// one Doublecast makes with the filter dropped: it waits for the next spell
// its controller casts this turn, whatever it is (and whether or not it has
// targets), and when it fires it copies that spell (copy created on the
// stack above it, CR 707.10, not cast, so nothing that triggers on a cast
// sees it), offering new targets for the copy. It triggers once (CR
// 603.7b) and ends at cleanup, copying nothing if no spell is cast. The
// original spell is copied from last-known information even if it was
// countered in response, as the ruling says.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e2d21635-182f-4ebf-a6f5-d4c1fd7f9705",
		Name:            "Flamehold Grappler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Flamehold Grappler — copy the next spell you cast this turn when you cast it",
				func(g *game.Game, item *game.StackItem) error {
					return WhenYouNextCast(
						"Flamehold Grappler — copy that spell; you may choose new targets for the copy",
						game.CastFilter{}, copyTheSpellBody,
					).Apply(NewContext(g, item))
				}),
		},
	})
}
