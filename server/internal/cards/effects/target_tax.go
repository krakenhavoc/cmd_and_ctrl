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
