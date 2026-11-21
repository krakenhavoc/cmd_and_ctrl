package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "9ad956c8-71d3-436f-961e-e2dfc1d2ca70",
		Name:            "Rust Harvester",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Exile an artifact card from your graveyard: Put a +1/+1 counter on this creature, then it deals damage equal to its power to any target.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.ExileFromGraveyard(1, "an artifact card", effects.MatchArtifact)),
			Targets: effects.TargetAny(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return effects.SourceDealsDamageEqualToItsPowerToTarget(g, item)
			},
		}},
	})
}
