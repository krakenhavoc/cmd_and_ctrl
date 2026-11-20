package effects

import "github.com/google/uuid"

// Stabbing Pain — Instant {B} (Magic 2011, #5579):
//
//	"Target creature gets -1/-1 until end of turn. Tap that creature."
//
// Rat Out's -1/-1 on a required target, then the tap, in printed order.
// The shrink is a layer 7c change that ends at cleanup; a creature it
// takes to 0 toughness is tapped and then dies to the state-based check
// once the spell has resolved (CR 704.5f). A target gone in response
// means the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "604ad5b3-7f11-455e-9442-65b7484238e5",
		Name:         "Stabbing Pain",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(HarmUntilEOT(0, -1, -1)),
		OnResolve: ResolveSteps(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
			if err := (BoostUntilEOT{Target: id, Power: -1, Toughness: -1, Label: "Stabbing Pain — -1/-1 until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			return TapTarget{Target: id}.Apply(ctx)
		})),
	})
}
