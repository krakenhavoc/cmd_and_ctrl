package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Recon Craft Theta — Artifact — Vehicle {4}, 4/4 (Fallout, #5948):
//
//	"Flying
//	 When this Vehicle enters, create a 0/0 blue Alien creature token. Put a
//	 +1/+1 counter on it.
//	 Whenever this Vehicle attacks, proliferate.
//	 Crew 2"
//
// The Alien is created and then given its counter in the same resolution,
// to the token the creation returned (a 0/0 that is not given the counter
// dies as a state-based action). The counter is not put on a token a
// replacement kept off the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8fba247e-9849-4a06-84ff-242894b0503a",
		Name:            "Recon Craft Theta",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Recon Craft Theta — create a 0/0 blue Alien creature token and put a +1/+1 counter on it",
				Do(Instruction(func(ctx *Context) error {
					ids, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(TokenCard("0/0 blue Alien")), 1, game.TokenEntryOptions{})
					if err != nil {
						return err
					}
					for _, id := range ids {
						if !onBattlefield(ctx.Game, id) {
							continue
						}
						if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}))),
			WheneverThisAttacks("Recon Craft Theta — proliferate", Do(Proliferate{})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Recon Craft Theta"),
		}},
	})
}
