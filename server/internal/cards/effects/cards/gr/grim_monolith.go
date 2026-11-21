package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grim Monolith —
//
// "This artifact doesn't untap during your untap step. {T}: Add {C}{C}{C}.
// {4}: Untap this artifact."
func init() {
	effects.Register(effects.Spec{
		OracleID:              "229d6627-1292-4ae1-8849-b0f956fa6540",
		Name:                  "Grim Monolith",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		ManaAbilities:         []effects.ManaAbility{{Cost: effects.ManaAbilityCost{Tap: true}, Produced: "{C}{C}{C}", Label: "Add {C}{C}{C}"}},
		Activated: []effects.ActivatedAbility{{
			// #2500: the whole effect is untapping the source.
			UntapSelf: true,
			Label:     "{4}: Untap Grim Monolith",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.ManaCost("{4}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.UntapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			}}},
	})
}
