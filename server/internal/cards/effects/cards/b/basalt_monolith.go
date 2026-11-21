package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Basalt Monolith —
//
// "This artifact doesn't untap during your untap step. {T}: Add {C}{C}{C}.
// {3}: Untap this artifact."
func init() {
	effects.Register(effects.Spec{
		OracleID:              "6b8cf2a0-b045-4d91-9d91-c602d40c6237",
		Name:                  "Basalt Monolith",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		ManaAbilities:         []effects.ManaAbility{{Cost: effects.ManaAbilityCost{Tap: true}, Produced: "{C}{C}{C}", Label: "Add {C}{C}{C}"}},
		Activated: []effects.ActivatedAbility{{
			// #2500: the whole effect is untapping the source.
			UntapSelf: true,
			Label:     "{3}: Untap Basalt Monolith",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.ManaCost("{3}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.UntapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			}}},
	})
}
