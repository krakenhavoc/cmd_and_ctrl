package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Contested Cliffs — Land (Onslaught):
//
//	"{T}: Add {C}.
//	 {R}{G}, {T}: Target Beast creature you control fights target
//	 creature an opponent controls."
//
// Ulvenwald Tracker's two-clause fight (slotZeroFightsSlotOne) on a
// land: slot 0 is a Beast you control, slot 1 a creature an opponent
// controls. If either target is illegal as the ability resolves,
// neither fights (CR 701.14b). A land has no summoning sickness, so the
// {T} may be paid the turn it is played.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "b891a683-2ebc-4e9c-b402-5dd9c1b42b69",
		Name:          "Contested Cliffs",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.PainlessColorless()},
		Activated: []effects.ActivatedAbility{{
			Label: "{R}{G}, {T}: Target Beast creature you control fights target creature an opponent controls.",
			Cost:  effects.Plus(effects.ManaCost("{R}{G}"), effects.TapCost()),
			Targets: effects.Clauses(
				effects.TargetCreature("target Beast creature you control", effects.YouControl(), effects.OfCreatureType("Beast")),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			),
			Effect: effects.SlotZeroFightsSlotOne,
		}},
	})
}
