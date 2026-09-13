package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deserted Temple — Land (EDHREC rank 1603):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Untap target land."
//
// The Cabal Coffers enabler. The untap is an ordinary CR 602
// activated ability with a mana-and-tap cost and a battlefield land
// target — any land, anyone's, as printed — and it uses the stack, so
// it can be responded to. The mana ability is the plain colourless
// tap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f12bf9a-6e1a-4377-b4af-e8cabd3ee58a",
		Name:         "Deserted Temple",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Untap target land.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetPermanent("target land", Land()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
