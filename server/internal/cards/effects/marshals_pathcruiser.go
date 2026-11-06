package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marshals' Pathcruiser — Artifact — Vehicle {3}, 6/5 (Aetherdrift,
// #4464):
//
//	"When this Vehicle enters, search your library for a basic land card,
//	 reveal it, put it into your hand, then shuffle.
//	 Exhaust — {W}{U}{B}{R}{G}: This Vehicle becomes an artifact creature.
//	 Put two +1/+1 counters on it. (Activate each exhaust ability only
//	 once.)
//	 Crew 5"
//
// The enters trigger is Horizon Seeker's basic-land search to hand
// (CR 701.23a), revealed; the search is not optional, though it may find
// nothing. The exhaust is Invasion Submersible's (CR 702.177a): the
// animation has no duration (CR 611.2a), so the Vehicle stays an artifact
// creature with its printed 6/5 (CR 301.7b) and the counters until it
// leaves; a Vehicle that left and came back in response is a new object
// and gets neither (CR 400.7). Crew is the engine's (CR 702.122a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "60f20f32-b6db-4ee7-a0aa-535647e6fcd5",
		Name:         "Marshals' Pathcruiser",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Marshals' Pathcruiser — search your library for a basic land card",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Marshals' Pathcruiser — a basic land card",
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "Exhaust — {W}{U}{B}{R}{G}: This Vehicle becomes an artifact creature. Put two +1/+1 counters on it.",
				Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerAnimate},
				Exhaust: true,
				Cost:    ManaCost("{W}{U}{B}{R}{G}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if sourceIsNewObject(g, item) {
						return nil
					}
					ctx := NewContext(g, item)
					if err := (BecomeArtifactCreature{
						Label: "Marshals' Pathcruiser — becomes an artifact creature",
					}).Apply(ctx); err != nil {
						return err
					}
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
				},
			},
			{
				Label:   "Crew 5",
				Cost:    CrewCost(5),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("Marshals' Pathcruiser"),
			},
		},
	})
}
