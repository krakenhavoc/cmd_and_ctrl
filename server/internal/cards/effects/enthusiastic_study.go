package effects

import "github.com/google/uuid"

// Enthusiastic Study — Instant {2}{R} (EDHREC rank 20291):
//
//	"Target creature gets +3/+1 and gains trample until end of turn.
//	 Learn. (You may reveal a Lesson card you own from outside the game
//	 and put it into your hand, or discard a card to draw a card.)"
//
// Both halves of the boost are one effect on the one creature, until end
// of turn (CR 611.2). Learn is the last instruction (CR 701.48a,
// learn.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4233334d-bb19-41e4-a203-2b7bb0c48f52",
		Name:         "Enthusiastic Study",
		Purpose:      ForTargets(BenefitUntilEOT(0, 3, 1, "trample")),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				if err := (BoostUntilEOT{
					Target: id, Power: 3, Toughness: 1,
					Label: "Enthusiastic Study — +3/+1 until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{
					Target:   id,
					Keywords: []string{"trample"},
					Label:    "Enthusiastic Study — trample until end of turn",
				}.Apply(ctx)
			}),
			Learn{},
		),
	})
}
