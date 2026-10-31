package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ultron Drone — {3} Artifact Creature — Robot Villain 2/3 (Marvel Super
// Heroes, #4455):
//
//	"Power-up — {6}: Put two +1/+1 counters on this creature and create a
//	 2/2 colorless Robot Villain artifact creature token. (Activate each
//	 power-up ability only once. Reduce the cost by its mana cost if it
//	 entered this turn.)"
//
// Pet Avengers' power-up shape: the counters on the source, then the token,
// as one resolution in printed order. Power-up is the engine's (CR 702.193).
//
// No simplification.
func init() {
	ab := PowerUp("{6}: Put two +1/+1 counters on this creature and create a 2/2 colorless Robot Villain artifact creature token.",
		ManaCost("{6}"), func(g *game.Game, item *game.StackItem) error {
			if err := plusOneCountersOnThis(2)(g, item); err != nil {
				return err
			}
			return CreateToken{Template: TokenCard("2/2 colorless Robot Villain artifact"), N: 1}.Apply(NewContext(g, item))
		})
	ab.Purpose.Answers = game.AnswerPump
	Register(Spec{
		OracleID:     "51c1f21d-1512-4885-bfdc-58790cf77d51",
		Name:         "Ultron Drone",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{ab},
	})
}
