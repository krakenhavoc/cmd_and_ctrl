package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Regal Caracal — Creature — Cat {3}{W}{W}, 3/3 (EDHREC rank 2988):
//
//	"Other Cats you control get +1/+1 and have lifelink. (Damage
//	 dealt by those creatures also causes you to gain that much
//	 life.)
//	 When this creature enters, create two 1/1 white Cat creature
//	 tokens with lifelink."
//
// The Cat lord that brings its own Cats. The two grants are
// TribalAnthem and TribalKeywordGrant over OTHER Cats the controller
// controls (both clauses printed); the enters trigger makes two
// b28WhiteCatLifelinkToken, which are Cats and so arrive as 2/2
// lifelinkers under the Caracal.
//
// No simplification.
func init() {
	cats := TribeFilter{Tribes: []string{"Cat"}, Others: true, YoursOnly: true}
	Register(Spec{
		OracleID:     "aa571676-b390-4b8a-baea-4c4cd60c01f6",
		Name:         "Regal Caracal",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(cats, 1, 1),
			TribalKeywordGrant(cats, "lifelink"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Regal Caracal — create two 1/1 white Cat tokens with lifelink", Do(CreateToken{Template: b28WhiteCatLifelinkToken(), N: 2})),
		},
	})
}
