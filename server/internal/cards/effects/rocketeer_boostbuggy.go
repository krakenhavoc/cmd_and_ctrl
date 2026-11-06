package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rocketeer Boostbuggy — Artifact — Vehicle {R}{G}, 3/2 (Aetherdrift,
// #4464):
//
//	"Whenever this Vehicle attacks, create a Treasure token. (It's an
//	 artifact with "{T}, Sacrifice this token: Add one mana of any
//	 color.")
//	 Exhaust — {3}: This Vehicle becomes an artifact creature. Put a
//	 +1/+1 counter on it. (Activate each exhaust ability only once.)
//	 Crew 1"
//
// The attack trigger makes one Treasure (CR 111.10a); the Vehicle can
// only attack while it is a creature. The exhaust is Marshals'
// Pathcruiser's (CR 702.177a) with one counter: the animation states no
// duration (CR 611.2a), so it stays an artifact creature with its
// printed 3/2 (CR 301.7b) until it leaves, and a Vehicle that left and
// came back in response is a new object and gets neither (CR 400.7).
// Crew is the engine's (CR 702.122a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2219e636-74d0-4f82-8704-e6d66c5f7f09",
		Name:         "Rocketeer Boostbuggy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Rocketeer Boostbuggy — create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "Exhaust — {3}: This Vehicle becomes an artifact creature. Put a +1/+1 counter on it.",
				Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerAnimate},
				Exhaust: true,
				Cost:    ManaCost("{3}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if sourceIsNewObject(g, item) {
						return nil
					}
					ctx := NewContext(g, item)
					if err := (BecomeArtifactCreature{
						Label: "Rocketeer Boostbuggy — becomes an artifact creature",
					}).Apply(ctx); err != nil {
						return err
					}
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				},
			},
			{
				Label:   "Crew 1",
				Cost:    CrewCost(1),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("Rocketeer Boostbuggy"),
			},
		},
	})
}
