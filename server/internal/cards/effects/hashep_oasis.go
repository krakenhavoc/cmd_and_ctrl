package effects

import "github.com/google/uuid"

// Hashep Oasis — Land — Desert:
//
//	"{T}: Add {C}.
//	 {T}, Pay 1 life: Add {G}.
//	 {1}{G}{G}, {T}, Sacrifice a Desert: Target creature gets +3/+3 until
//	 end of turn. Activate only as a sorcery."
//
// No simplification.
func init() {
	Register(cmdr2B13DesertLand("eab70fff-6a9f-4f9f-89a2-b6910c199e46", "Hashep Oasis", "G", ActivatedAbility{
		Label:        "{1}{G}{G}, {T}, Sacrifice a Desert: Target creature gets +3/+3 until end of turn. Activate only as a sorcery.",
		Purpose:      ForTargets(BenefitUntilEOT(0, 3, 3)),
		Cost:         Plus(ManaCost("{1}{G}{G}"), TapCost(), cmdr2B13SacrificeADesert()),
		SorcerySpeed: true,
		Targets:      TargetCreature("target creature"),
		Effect: Do(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
			return BoostUntilEOT{Target: id, Power: 3, Toughness: 3,
				Label: "Hashep Oasis — +3/+3 until end of turn"}.Apply(ctx)
		})),
	}))
}
