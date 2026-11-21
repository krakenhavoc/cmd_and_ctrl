package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "2ac2b815-2d72-48e6-b43a-18884a74bf95",
		Name:         "Skycoach Waypoint",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Target creature becomes prepared.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					return effects.BecomePrepared{Target: t.ID}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
