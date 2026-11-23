package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Minas Morgul, Dark Fortress — Legendary Land:
//
//	"Minas Morgul enters tapped.
//	 {T}: Add {B}.
//	 {3}{B}, {T}: Put a shadow counter on target creature. For as long as that creature has a shadow counter on it, it's a Wraith in addition to its other types. (A creature with shadow can block or be blocked by only creatures with shadow.)"
//
// ADR 0109 §2 (#1604): a shadow counter is a keyword counter (CR
// 122.1b), so the creature has shadow while it has one, on its own.
// The Wraith type is the resolved effect: an addSubtypes record for as
// long as the creature has a shadow counter on it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "867dbd5a-c3cf-41ce-980b-c9babc6f30f2",
		Name:         "Minas Morgul, Dark Fortress",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}, {T}: Put a shadow counter on target creature. For as long as that creature has a shadow counter on it, it's a Wraith in addition to its other types. (A creature with shadow can block or be blocked by only creatures with shadow.)",
			Cost:    effects.Plus(effects.ManaCost("{3}{B}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.CounterThenWhileItHasIt(ctx, effects.FirstLegalBattlefieldTarget(ctx), game.CounterShadow,
					"Minas Morgul, Dark Fortress — a Wraith while it has a shadow counter", game.AddSubtypesMod(ctx.TextSubtypes("Wraith")...))
			},
		}},
	})
}
