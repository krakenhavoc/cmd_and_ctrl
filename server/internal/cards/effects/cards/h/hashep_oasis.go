package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hashep Oasis — Land — Desert:
//
//	"{T}: Add {C}.
//	 {T}, Pay 1 life: Add {G}.
//	 {1}{G}{G}, {T}, Sacrifice a Desert: Target creature gets +3/+3 until
//	 end of turn. Activate only as a sorcery."
//
// No simplification.
func init() {
	effects.Register(effects.Cmdr2B13DesertLand("eab70fff-6a9f-4f9f-89a2-b6910c199e46", "Hashep Oasis", "G", effects.ActivatedAbility{
		Label:        "{1}{G}{G}, {T}, Sacrifice a Desert: Target creature gets +3/+3 until end of turn. Activate only as a sorcery.",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3)),
		Cost:         effects.Plus(effects.ManaCost("{1}{G}{G}"), effects.TapCost(), effects.Cmdr2B13SacrificeADesert()),
		SorcerySpeed: true,
		Targets:      effects.TargetCreature("target creature"),
		Effect: effects.Do(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
			return effects.BoostUntilEOT{Target: id, Power: 3, Toughness: 3,
				Label: "Hashep Oasis — +3/+3 until end of turn"}.Apply(ctx)
		})),
	}))
}
