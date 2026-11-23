package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bestow Greatness — Instant {2}{G} (Aetherdrift, #4464):
//
//	"Target creature gets +4/+4 and gains trample until end of turn."
//
// Enthusiastic Study's body without the learn: one target, so a target
// that has gone by resolution counters the spell (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "58a17110-b2a7-4dbc-bba5-5ed3728ad261",
		Name:         "Bestow Greatness",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 4, 4, "trample")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				if err := (effects.BoostUntilEOT{
					Target: id, Power: 4, Toughness: 4,
					Label: "Bestow Greatness — +4/+4 until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{
					Target:   id,
					Keywords: []string{"trample"},
					Label:    "Bestow Greatness — trample until end of turn",
				}.Apply(ctx)
			}),
		),
	})
}
