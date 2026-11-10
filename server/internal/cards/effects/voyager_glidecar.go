package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Voyager Glidecar — Artifact — Vehicle {W}, 2/3 (Aetherdrift, #4464):
//
//	"When this Vehicle enters, scry 1.
//	 Tap three other untapped creatures you control: Until end of turn,
//	 this Vehicle becomes an artifact creature and gains flying. Put a
//	 +1/+1 counter on it.
//	 Crew 1"
//
// The tapping ability is a TapOthers cost for three creatures other than
// the Vehicle (not the {T} symbol, CR 302.6, so a creature that just
// arrived may pay) with the animation's own effect, the Vehicle becoming an
// artifact creature until end of turn and gaining flying, then the counter,
// which stays after the turn (a creature Vehicle is bound to the object, CR
// 400.7). Crew 1 is the ordinary crew row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f40d73b-9106-448e-8649-5d62939f0d38",
		Name:         "Voyager Glidecar",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Voyager Glidecar — scry 1", Do(Scry{N: 1})),
		},
		Activated: []ActivatedAbility{
			{
				Label: "Tap three other untapped creatures you control: Until end of turn, this Vehicle becomes an artifact creature and gains flying. Put a +1/+1 counter on it.",
				Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
					Count:         3,
					Filter:        TargetPermanent("three other untapped creatures you control", Creature()),
					ExcludeSource: true,
					Label:         "three other untapped creatures you control",
				}},
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  voyagerGlidecarAnimate,
			},
			{
				Label:   "Crew 1",
				Cost:    CrewCost(1),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("Voyager Glidecar"),
			},
		},
	})
}

func voyagerGlidecarAnimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (BecomeCreatureUntilEOT{Label: "Voyager Glidecar — artifact creature"}).Apply(ctx); err != nil {
		return err
	}
	if err := (GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"flying"},
		Label: "Voyager Glidecar — flying"}).Apply(ctx); err != nil {
		return err
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}
