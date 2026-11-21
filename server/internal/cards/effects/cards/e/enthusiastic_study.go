package cardse

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
	effects.Register(effects.Spec{
		OracleID:     "4233334d-bb19-41e4-a203-2b7bb0c48f52",
		Name:         "Enthusiastic Study",
		Purpose:      effects.WithRiders(effects.ForTargets(effects.BenefitUntilEOT(0, 3, 1, "trample")), game.Purpose{Learn: 1}),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				if err := (effects.BoostUntilEOT{
					Target: id, Power: 3, Toughness: 1,
					Label: "Enthusiastic Study — +3/+1 until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{
					Target:   id,
					Keywords: []string{"trample"},
					Label:    "Enthusiastic Study — trample until end of turn",
				}.Apply(ctx)
			}),
			effects.Learn{},
		),
	})
}
