package effects

import "github.com/google/uuid"

// Toxic Stench — Instant {1}{B} (Judgment, #5611):
//
//	"Target nonblack creature gets -1/-1 until end of turn.
//	 Threshold — If there are seven or more cards in your graveyard,
//	 instead destroy that creature. It can't be regenerated."
//
// Threshold is read as the spell resolves (CR 207.2c), with the spell on
// the stack and so not one of the seven. Below it, the -1/-1 ends at
// cleanup (CR 514.2); at it, the shrink is replaced by a destruction
// that ignores regeneration shields (CR 701.19c). Declared as a harm.
// A target that has gone, or has become black, counters the spell
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9c3be33b-0be9-49ec-8160-898cbcab8814",
		Name:         "Toxic Stench",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(HarmUntilEOT(0, -1, -1)),
		Targets:      TargetCreature("target nonblack creature", NonBlack()),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				if fraThresholdMet(ctx.Game, ctx.Controller()) {
					return DestroyTarget{Target: id, CantBeRegenerated: true}.Apply(ctx)
				}
				return BoostUntilEOT{Target: id, Power: -1, Toughness: -1, Label: "Toxic Stench — -1/-1"}.Apply(ctx)
			}),
		),
	})
}
