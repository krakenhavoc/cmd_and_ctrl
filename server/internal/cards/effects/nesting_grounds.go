package effects

import (
	"sort"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nesting Grounds — Land:
//
//	"{T}: Add {C}.
//	 {1}, {T}: Move a counter from target permanent you control onto a
//	 second target permanent. Activate only as a sorcery."
//
// Two target clauses, the second Distinct (CR 122.5 says moving a
// counter from an object onto itself is not possible, and "a second
// target" means a different one). At resolution both must still be
// legal targets (CR 608.2b). When the first permanent carries one kind
// of counter that kind moves; when it carries several, the controller
// is asked which (PickOption), since the printed text lets them choose.
//
// The move is "put it on the second permanent, then remove it from the
// first only if it landed" (CR 122.5), the same order Simic Fluxmage
// uses, so a counter the second permanent can't take is not lost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d27bb97d-286b-4947-8d7b-443e4df93319",
		Name:         "Nesting Grounds",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{1}, {T}: Move a counter from target permanent you control onto a second target permanent. Activate only as a sorcery.",
			Cost:  Plus(ManaCost("{1}"), TapCost()),
			Targets: Clauses(
				TargetPermanent("target permanent you control", YouControl()),
				Distinct(TargetPermanent("a second target permanent")),
			),
			SorcerySpeed: true,
			Effect:       nestingGroundsMove,
		}},
	})
}

func nestingGroundsMove(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	from, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	to, ok := ctx.ClauseTarget(1)
	if !ok || from.ID == to.ID {
		return nil
	}
	src, ok := g.LookupCardForEffect(from.ID)
	if !ok {
		return nil
	}
	var kinds []string
	for kind, n := range src.Counters {
		if n > 0 {
			kinds = append(kinds, kind)
		}
	}
	sort.Strings(kinds)
	move := func(kind string) error {
		fromID, toID := from.ID, to.ID
		if _, ok := g.LookupCardForEffect(toID); !ok {
			return nil
		}
		return g.AddCounterByThenForEffect(item.Controller, toID, kind, 1, func(g *game.Game, placed int) error {
			if placed <= 0 {
				return nil
			}
			return g.AddCounterForEffect(fromID, kind, -1)
		})
	}
	switch len(kinds) {
	case 0:
		return nil
	case 1:
		return move(kinds[0])
	}
	options := make([]game.ChoiceOption, len(kinds))
	for i, kind := range kinds {
		options[i] = game.ChoiceOption{Label: "Move a " + kind + " counter"}
	}
	return PickOption{
		Question: "Nesting Grounds — which kind of counter do you move?",
		Options:  options,
		Then: func(_ *Context, index int) error {
			if index < 0 || index >= len(kinds) {
				return nil
			}
			return move(kinds[index])
		},
	}.Apply(ctx)
}
