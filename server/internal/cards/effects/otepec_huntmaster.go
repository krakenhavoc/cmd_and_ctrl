package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "41623eab-38b5-4058-8de2-e6b8706de7a2",
		Name:         "Otepec Huntmaster",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Dinosaur spells you cast cost {1} less to cast.",
				YourSpell(), SpellOfCreatureType("Dinosaur")),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target Dinosaur gains haste until end of turn.",
			Cost:    TapCost(),
			Targets: TargetCreature("target Dinosaur", Subtype("Dinosaur")),
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "haste")),
			Effect: Do(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: label}.Apply(ctx)
			})),
		}},
	})
}
