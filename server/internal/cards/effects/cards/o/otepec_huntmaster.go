package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Otepec Huntmaster — Creature — Human Shaman {1}{R}, 1/2 (Lost Caverns
// of Ixalan Commander):
//
//	"Dinosaur spells you cast cost {1} less to cast.
//	 {T}: Target Dinosaur gains haste until end of turn."
//
// A cost modifier (CR 601.2f) over the controller's Dinosaur spells and a
// tap ability whose target is any Dinosaur on the battlefield, read through
// the layered view so a changeling counts.
func init() {
	const label = "Otepec Huntmaster — gains haste until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "41623eab-38b5-4058-8de2-e6b8706de7a2",
		Name:         "Otepec Huntmaster",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Dinosaur spells you cast cost {1} less to cast.",
				effects.YourSpell(), effects.SpellOfCreatureType("Dinosaur")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target Dinosaur gains haste until end of turn.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target Dinosaur", effects.Subtype("Dinosaur")),
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
			Effect: effects.Do(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: label}.Apply(ctx)
			})),
		}},
	})
}
