package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// splice.go — CR 702.47, splice (#1989, ADR 0202). The card side is one
// field, Spec.Splice; the engine reveals, prices, targets and runs the
// text (game/splice.go).

// SpliceOntoArcane is "Splice onto Arcane [cost]".
func SpliceOntoArcane(manaCost string) *game.SpliceAbility {
	return &game.SpliceAbility{Onto: game.SpliceOntoArcane, ManaCost: manaCost}
}

// SpliceOntoInstantOrSorcery is "Splice onto instant or sorcery [cost]".
func SpliceOntoInstantOrSorcery(manaCost string) *game.SpliceAbility {
	return &game.SpliceAbility{Onto: game.SpliceOntoInstantOrSorcery, ManaCost: manaCost}
}

// checkSplice refuses a splice declaration the engine cannot carry out
// as printed:
//
//   - no cost, or a cost that is not mana: a non-mana splice cost
//     (Veil of Secrecy's "Return a blue creature you control to its
//     owner's hand") is not built, and a card without its cost would
//     be stronger than printed;
//   - an unknown quality;
//   - a text the engine could not add to another spell whole: no body,
//     a body split into modes, an announced X, a division, or a cost of
//     its own besides the splice cost. CR 702.47c adds the RULES TEXT,
//     and none of those is read off a spell's spliced text.
func checkSplice(spec Spec) {
	sa := spec.Splice
	if sa == nil {
		return
	}
	switch sa.Onto {
	case game.SpliceOntoArcane, game.SpliceOntoInstantOrSorcery:
	default:
		panic(fmt.Sprintf("effects.Register: %q splices onto an unknown quality %q", spec.Name, sa.Onto))
	}
	if sa.ManaCost == "" {
		panic(fmt.Sprintf("effects.Register: %q declares a splice cost that demands nothing — a non-mana splice cost is not built (ADR 0202)", spec.Name))
	}
	if _, err := game.ParseCost(sa.ManaCost); err != nil {
		panic(fmt.Sprintf("effects.Register: %q splice cost %q: %v", spec.Name, sa.ManaCost, err))
	}
	if spec.OnResolve == nil {
		panic(fmt.Sprintf("effects.Register: %q declares splice but no OnResolve — its rules text is what a spliced spell gains", spec.Name))
	}
	if spec.Modes != nil || spec.AdditionalCost != nil || len(spec.OptionalCosts) > 0 || spec.XCeiling != nil {
		panic(fmt.Sprintf("effects.Register: %q declares splice beside modes or a cost of its own — a spliced text carries neither (ADR 0202)", spec.Name))
	}
	if spec.Targets != nil {
		for _, c := range spec.Targets.Clauses() {
			if c.CountFromX || c.Divide != nil {
				panic(fmt.Sprintf("effects.Register: %q declares splice with an X- or division-bound target — the spliced spell announces neither for it (ADR 0202)", spec.Name))
			}
		}
	}
}
