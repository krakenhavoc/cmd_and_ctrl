package effects

import "github.com/google/uuid"

// Arcane Subtraction — Instant {1}{U} (EDHREC rank 25602):
//
//	"Target creature gets -4/-0 until end of turn.
//	 Learn. (You may reveal a Lesson card you own from outside the game
//	 and put it into your hand, or discard a card to draw a card.)"
//
// A negative boost is how a shrink is written (BoostUntilEOT); power is
// clamped at zero where damage is read, toughness is not touched. Learn
// is the last instruction (CR 701.48a, learn.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b5806fa-430f-4b78-b874-63d01a8a9de5",
		Name:         "Arcane Subtraction",
		Purpose:      ForTargets(HarmUntilEOT(0, -4, 0)),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return BoostUntilEOT{
					Target: id, Power: -4, Toughness: 0,
					Label: "Arcane Subtraction — -4/-0 until end of turn",
				}.Apply(ctx)
			}),
			Learn{},
		),
	})
}
