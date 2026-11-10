package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tangle Tumbler — Artifact — Vehicle {3}, 6/6 (Bloomburrow, #4467):
//
//	"Vigilance
//	 {3}, {T}: Put a +1/+1 counter on target creature.
//	 Tap two untapped tokens you control: This Vehicle becomes an artifact
//	 creature until end of turn."
//
// The counter ability taps the Vehicle, which is not a creature until
// animated, so it can be activated the turn it arrives (CR 302.6 asks only
// of a creature). The animation is crew's effect (BecomeCreatureUntilEOT)
// behind a TapOthers cost of two untapped tokens you control: any tokens,
// creature or not, summoning-sick or not (the ruling: it is not the {T}
// symbol), and nothing that cares about being crewed sees it (CR 702.122e).
// Vigilance rides PrintedKeywords.
//
// No simplification.
func init() {
	const twoTokens = "two untapped tokens you control"
	Register(Spec{
		OracleID:        "0d81ea20-5e7d-4eea-b0ba-fe688f27df94",
		Name:            "Tangle Tumbler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{
			{
				Label:   "{3}, {T}: Put a +1/+1 counter on target creature.",
				Cost:    Plus(ManaCost("{3}"), TapCost()),
				Targets: TargetCreature("target creature"),
				Effect:  PlusOneCounterOnTheTarget,
			},
			{
				Label: "Tap two untapped tokens you control: This Vehicle becomes an artifact creature until end of turn.",
				Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
					Count:  2,
					Filter: TargetPermanent(twoTokens, IsTokenPredicate()),
					Label:  twoTokens,
				}},
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return BecomeCreatureUntilEOT{Label: "Tangle Tumbler — becomes an artifact creature"}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
