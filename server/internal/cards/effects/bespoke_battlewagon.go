package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bespoke Battlewagon — Artifact — Vehicle {3}{U}, 5/6:
//
//	"{T}: You get {E}{E} (two energy counters).
//	 {T}, Pay {E}{E}: Tap target creature.
//	 {T}, Pay {E}{E}{E}: Draw a card.
//	 Pay {E}{E}{E}{E}: This Vehicle becomes an artifact creature until
//	 end of turn.
//	 Crew 4"
//
// ADR 0129 PR 1 (#1995). The Vehicle is not a creature until crewed or
// animated, so its {T} abilities are not held back by summoning
// sickness (CR 302.6) until it is one. "Becomes an artifact creature"
// is the same until-end-of-turn type change crewing makes, paid with
// energy instead of tapping creatures.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cdfcfd2b-966a-4e03-8ed5-4fa7fab423d0",
		Name:         "Bespoke Battlewagon",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: You get {E}{E}.",
				Cost:    TapCost(),
				Purpose: game.Purpose{Answers: game.AnswerValue, Energy: 2},
				Effect:  ebYouGetEnergy(2),
			},
			{
				Label:   "{T}, Pay {E}{E}: Tap target creature.",
				Cost:    Plus(TapCost(), PayEnergy(2)),
				Targets: TargetCreature("target creature"),
				Effect:  tapChosenPermanent,
			},
			{
				Label:   "{T}, Pay {E}{E}{E}: Draw a card.",
				Cost:    Plus(TapCost(), PayEnergy(3)),
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Effect:  ebDrawACard,
			},
			{
				Label: "Pay {E}{E}{E}{E}: This Vehicle becomes an artifact creature until end of turn.",
				Cost:  PayEnergy(4),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return BecomeCreatureUntilEOT{Label: "Bespoke Battlewagon — becomes an artifact creature"}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "Crew 4",
				Cost:    CrewCost(4),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("Bespoke Battlewagon"),
			},
		},
	})
}
