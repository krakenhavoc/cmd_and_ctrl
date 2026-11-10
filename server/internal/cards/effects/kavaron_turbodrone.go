package effects

import "github.com/google/uuid"

// Kavaron Turbodrone — Artifact Creature — Robot Scout {2}{R}, 2/3 (Edge
// of Eternities, #4461):
//
//	"{T}: Target creature you control gets +1/+1 and gains haste until
//	 end of turn. Activate only as a sorcery."
//
// Cunning Coyote's +1/+1 and haste on an activated row with Captain
// Jonathan Archer's sorcery-speed tap. The Turbodrone may target itself.
// A target gone by resolution makes the ability do nothing (CR 608.2b).
// Sorcery speed, so the row declares no Answers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ebb8542-5d1d-4f0b-92a9-5a98425860fd",
		Name:         "Kavaron Turbodrone",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{T}: Target creature you control gets +1/+1 and gains haste until end of turn. Activate only as a sorcery.",
			Purpose:      ForTargets(BenefitUntilEOT(0, 1, 1, "haste")),
			Cost:         TapCost(),
			Targets:      TargetCreature("target creature you control", YouControl()),
			SorcerySpeed: true,
			Effect: Do(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				if err := (BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Kavaron Turbodrone — +1/+1"}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Kavaron Turbodrone — haste"}.Apply(ctx)
			})),
		}},
	})
}
