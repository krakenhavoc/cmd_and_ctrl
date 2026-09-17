package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grim Monolith —
//
// "This artifact doesn't untap during your untap step. {T}: Add {C}{C}{C}.
// {4}: Untap this artifact."
func init() {
	Register(Spec{
		OracleID:              "229d6627-1292-4ae1-8849-b0f956fa6540",
		Name:                  "Grim Monolith",
		Completeness:          CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{doesntUntapDuringYourUntapStep()},
		ManaAbilities:         []ManaAbility{{Cost: ManaAbilityCost{Tap: true}, Produced: "{C}{C}{C}", Label: "Add {C}{C}{C}"}},
		Activated: []ActivatedAbility{{
			Label: "{4}: Untap Grim Monolith",
			Cost:  ManaCost("{4}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return UntapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
			}}},
	})
}
