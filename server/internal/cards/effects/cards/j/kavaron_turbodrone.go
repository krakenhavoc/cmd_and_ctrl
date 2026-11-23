package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "3ebb8542-5d1d-4f0b-92a9-5a98425860fd",
		Name:         "Kavaron Turbodrone",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}: Target creature you control gets +1/+1 and gains haste until end of turn. Activate only as a sorcery.",
			Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "haste")),
			Cost:         effects.TapCost(),
			Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
			SorcerySpeed: true,
			Effect: effects.Do(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				if err := (effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Kavaron Turbodrone — +1/+1"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Kavaron Turbodrone — haste"}.Apply(ctx)
			})),
		}},
	})
}
