package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skycoach Waypoint — Land:
//
//	"{T}: Add {C}.
//	 {3}, {T}: Target creature becomes prepared. (Only creatures with
//	 prepare spells can become prepared.)"
//
// The "target creature becomes prepared" proof card for ADR 0090, and
// the one that exercises CR 722.3a's refusals from outside the card
// being prepared: the ability targets ANY creature, as printed, and
// resolves into nothing on a creature with no prepare spell or one
// that is already prepared. The reminder text is the rule, not a
// targeting restriction — narrowing the target clause to preparation
// cards would make a legal activation illegal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2ac2b815-2d72-48e6-b43a-18884a74bf95",
		Name:         "Skycoach Waypoint",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: Target creature becomes prepared.",
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					return BecomePrepared{Target: t.ID}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
