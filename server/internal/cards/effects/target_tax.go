package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// target_tax.go — constructors for Spec.TargetTaxes, the cost a
// permanent adds to spells that TARGET something (CR 601.2f, #3860).
// The engine half lives in game/target_tax.go; this file is the
// vocabulary a card file writes in.
//
// The shape to copy (Terror of the Peaks):
//
//	TargetTaxes: []game.TargetTax{
//	    OpponentsSpellsTargetingThisCostLife(3,
//	        "Spells your opponents cast that target this creature cost an additional 3 life to cast."),
//	},
//
// The label is the printed clause, taken rather than generated, for
// attack_tax.go's reason: the refusal and the picker's note quote it.
//
// "YOUR opponents" is structural -- the caster is anyone but the
// permanent's controller -- and "this creature" is the permanent itself,
// so a card file cannot widen either by accident.

// OpponentsSpellsTargetingThisCostLife is "Spells your opponents cast
// that target this <permanent> cost an additional N life to cast": the
// caster pays `life` once, however many of the spell's targets name the
// permanent. The permanent's own controller's spells are free, and so is
// an ability that targets it (that is ward's business) and a copy of a
// spell (it is not cast, CR 707.10).
func OpponentsSpellsTargetingThisCostLife(life int, label string) game.TargetTax {
	return game.TargetTax{
		Label: label,
		Life:  life,
		AppliesTo: func(q game.TargetTaxQuery) bool {
			return q.Target.InstanceID == q.Source.InstanceID && q.Caster != q.Source.Controller
		},
	}
}

// CostsLifePerTarget is the spell's OWN "This spell costs N life more
// to cast for each target" (Phyrexian Purge, #6206, CR 601.2f): the
// caster pays `life` for every target the spell names, checked against
// their life total with the cast's other life costs (CR 119.4) and paid
// with the rest of its additional costs (CR 601.2h).
//
// It goes in Spec.LifePerTarget rather than Spec.TargetTaxes, because
// the clause is the card's own: no permanent charges it, so no CR 613.1f
// ability-removing effect on the battlefield can take it away.
func CostsLifePerTarget(life int, label string) game.LifePerTargetCost {
	return game.LifePerTargetCost{Life: life, Label: label}
}

// checkLifePerTarget is Register's guard for Spec.LifePerTarget — the
// bargain checkTargetTaxes makes, one slot over: a clause with no
// printed Label gives a refusal and a picker note nobody can read, and
// one that charges nothing is dead text claiming a rule.
func checkLifePerTarget(card string, c game.LifePerTargetCost) {
	if c.Life == 0 && c.Label == "" {
		return
	}
	if c.Life <= 0 {
		panic(fmt.Sprintf("effects.Register: %q per-target life cost %q charges no life", card, c.Label))
	}
	if c.Label == "" {
		panic(fmt.Sprintf("effects.Register: %q per-target life cost has no printed Label -- the refusal and the picker quote it", card))
	}
}

// checkTargetTaxes is Register's guard for Spec.TargetTaxes: a tax with
// no printed Label produces a refusal and a picker note nobody can read,
// one that charges nothing is dead text claiming a rule, and one with no
// AppliesTo would tax every spell on the table.
func checkTargetTaxes(card string, ts []game.TargetTax) {
	for i, t := range ts {
		if t.Label == "" {
			panic(fmt.Sprintf("effects.Register: %q target tax %d has no printed Label -- the refusal and the picker quote it", card, i))
		}
		if t.Life <= 0 {
			panic(fmt.Sprintf("effects.Register: %q target tax %q charges no life", card, t.Label))
		}
		if t.AppliesTo == nil {
			panic(fmt.Sprintf("effects.Register: %q target tax %q names no spell and no target", card, t.Label))
		}
	}
}
