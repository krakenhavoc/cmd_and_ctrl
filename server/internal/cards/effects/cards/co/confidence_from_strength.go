package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Confidence from Strength — Sorcery {2}{G} (Global Series Jiang Yanggu & Mu Yanling, S82 OLD batch 201, #5528):
//
//	"Target creature gets +4/+4 and gains trample until end of turn."
//
// Larger Than Life's pump and trample grant, each a continuous effect until
// end of turn on the target as it resolves (CR 611.2c). With its one target
// illegal the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "341f75a5-df02-48e9-8997-1938148ad89a",
		Name:         "Confidence from Strength",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 4, 4, "trample")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				if err := (effects.BoostUntilEOT{
					Target: id, Power: 4, Toughness: 4,
					Label: "Confidence from Strength — +4/+4 until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{
					Target:   id,
					Keywords: []string{"trample"},
					Label:    "Confidence from Strength — trample until end of turn",
				}.Apply(ctx)
			}),
		),
	})
}
