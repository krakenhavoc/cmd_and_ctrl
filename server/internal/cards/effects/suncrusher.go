package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Suncrusher — Artifact Creature — Construct {9}, 3/3:
//
//	"Sunburst (This creature enters with a +1/+1 counter on it for each
//	 color of mana spent to cast it.)
//	 {4}, {T}, Remove a +1/+1 counter from this creature: Destroy
//	 target creature.
//	 {2}, Remove a +1/+1 counter from this creature: Return this
//	 creature to its owner's hand."
//
// Sunburst is a keyword the engine reads off the resolving spell (ADR
// 0109 §11, #1552). Both abilities pay with the counters it made.
func init() {
	Register(Spec{
		OracleID:            "4e81ead0-df8b-4d9d-b1cb-d1d9dc503d4c",
		Name:                "Suncrusher",
		Completeness:        CompletenessCaveats,
		Caveats:             []string{"With strict mana off, the game doesn't track which mana you spent, so it enters with no counters at all."},
		WantsDistinctColors: true,
		PrintedKeywords:     []string{game.KeywordSunburst},
		Activated: []ActivatedAbility{
			{
				Label:   "{4}, {T}, Remove a +1/+1 counter from this creature: Destroy target creature.",
				Cost:    Plus(ManaCost("{4}"), TapCost(), RemoveCountersFromThis(game.CounterPlusOne, 1)),
				Targets: TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return destroyTheTargetPermanent(item, NewContext(g, item))
				},
			},
			{
				Label:   "{2}, Remove a +1/+1 counter from this creature: Return this creature to its owner's hand.",
				Purpose: game.Purpose{Answers: game.AnswerProtect},
				Cost:    Plus(ManaCost("{2}"), RemoveCountersFromThis(game.CounterPlusOne, 1)),
				Effect:  returnThisPermanentToOwnersHand,
			},
		},
	})
}
