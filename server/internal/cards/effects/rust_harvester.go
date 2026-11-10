package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rust Harvester — Artifact Creature — Robot {R}, 1/1 (Edge of
// Eternities):
//
//	"Menace
//	 {2}, {T}, Exile an artifact card from your graveyard: Put a +1/+1
//	 counter on this creature, then it deals damage equal to its power to
//	 any target."
//
// The mana, the tap and the exiled artifact card are all costs (CR
// 602.2). The counter goes on first, so the damage counts it; a Harvester
// that has left the battlefield gets no counter and deals damage equal to
// its power as it last existed (CR 608.2h, the rulings). A target that is
// illegal on resolution fizzles the whole ability, counter included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9ad956c8-71d3-436f-961e-e2dfc1d2ca70",
		Name:            "Rust Harvester",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Exile an artifact card from your graveyard: Put a +1/+1 counter on this creature, then it deals damage equal to its power to any target.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    Plus(ManaCost("{2}"), TapCost(), ExileFromGraveyard(1, "an artifact card", MatchArtifact)),
			Targets: TargetAny(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (CounterOnThis{Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return sourceDealsDamageEqualToItsPowerToTarget(g, item)
			},
		}},
	})
}
