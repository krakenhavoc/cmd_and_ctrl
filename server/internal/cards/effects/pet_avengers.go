package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pet Avengers — {3}{G} Creature — Dragon Cat Dog Bird Frog Hero 4/4
// (Marvel Super Heroes, #4455):
//
//	"Reach
//	 Power-up — {6}{G}: Put a +1/+1 counter on this creature and create a
//	 3/2 white Hero creature token with vigilance. (Activate each power-up
//	 ability only once. Reduce the cost by its mana cost if it entered this
//	 turn.)"
//
// Power-up is the engine's (CR 702.193, game/power_up.go): the turn it
// enters the ability costs {6}{G} less {3}{G}, which is {3}; after that
// {6}{G}. The counter and the token are one resolution; the counter is
// skipped if the creature has left, the token is not.
//
// No simplification.
func init() {
	ab := PowerUp("{6}{G}: Put a +1/+1 counter on this creature and create a 3/2 white Hero creature token with vigilance.",
		ManaCost("{6}{G}"), func(g *game.Game, item *game.StackItem) error {
			if err := plusOneCountersOnThis(1)(g, item); err != nil {
				return err
			}
			return CreateToken{Template: TokenCard("3/2 white Hero with vigilance"), N: 1}.Apply(NewContext(g, item))
		})
	ab.Purpose.Answers = game.AnswerPump
	Register(Spec{
		OracleID:        "bbee7921-19c9-4c70-8462-f16e7e9c25db",
		Name:            "Pet Avengers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated:       []ActivatedAbility{ab},
	})
}
