package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Basalt Monolith —
//
// "This artifact doesn't untap during your untap step. {T}: Add {C}{C}{C}.
// {3}: Untap this artifact."
func init() {
	Register(Spec{
		OracleID:              "6b8cf2a0-b045-4d91-9d91-c602d40c6237",
		Name:                  "Basalt Monolith",
		Completeness:          CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{doesntUntapDuringYourUntapStep()},
		ManaAbilities:         []ManaAbility{{Cost: ManaAbilityCost{Tap: true}, Produced: "{C}{C}{C}", Label: "Add {C}{C}{C}"}},
		Activated: []ActivatedAbility{{
			Label: "{3}: Untap Basalt Monolith",
			Cost:  ManaCost("{3}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return UntapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
			}}},
	})
}
