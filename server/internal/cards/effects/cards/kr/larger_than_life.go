package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Larger Than Life — Sorcery {1}{G} (Kaladesh, #5539):
//
//	"Target creature gets +4/+4 and gains trample until end of turn."
//
// Bestow Greatness's body: a layer 7c +4/+4 and a layer 6 trample,
// both until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "96e8210a-a547-431d-8146-7399ec20c591",
		Name:         "Larger Than Life",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 4, 4, "trample")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				if err := (effects.BoostUntilEOT{
					Target: id, Power: 4, Toughness: 4,
					Label: "Larger Than Life — +4/+4 until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{
					Target:   id,
					Keywords: []string{"trample"},
					Label:    "Larger Than Life — trample until end of turn",
				}.Apply(ctx)
			}),
		),
	})
}
