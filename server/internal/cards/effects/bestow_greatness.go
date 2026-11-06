package effects

import "github.com/google/uuid"

// Bestow Greatness — Instant {2}{G} (Aetherdrift, #4464):
//
//	"Target creature gets +4/+4 and gains trample until end of turn."
//
// Enthusiastic Study's body without the learn: one target, so a target
// that has gone by resolution counters the spell (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "58a17110-b2a7-4dbc-bba5-5ed3728ad261",
		Name:         "Bestow Greatness",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 4, 4, "trample")),
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				if err := (BoostUntilEOT{
					Target: id, Power: 4, Toughness: 4,
					Label: "Bestow Greatness — +4/+4 until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{
					Target:   id,
					Keywords: []string{"trample"},
					Label:    "Bestow Greatness — trample until end of turn",
				}.Apply(ctx)
			}),
		),
	})
}
