package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "867dbd5a-c3cf-41ce-980b-c9babc6f30f2",
		Name:         "Minas Morgul, Dark Fortress",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}{B}, {T}: Put a shadow counter on target creature. For as long as that creature has a shadow counter on it, it's a Wraith in addition to its other types. (A creature with shadow can block or be blocked by only creatures with shadow.)",
			Cost:    Plus(ManaCost("{3}{B}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return CounterThenWhileItHasIt(ctx, FirstLegalBattlefieldTarget(ctx), game.CounterShadow,
					"Minas Morgul, Dark Fortress — a Wraith while it has a shadow counter", game.AddSubtypesMod("Wraith"))
			},
		}},
	})
}
