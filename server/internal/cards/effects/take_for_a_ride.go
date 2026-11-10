package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Take for a Ride — Sorcery {2}{R} (Outlaws of Thunder Junction):
//
//	"Take for a Ride has flash as long as you've committed a crime this turn.
//	 (Targeting opponents, anything they control, and/or cards in their
//	 graveyards is a crime.)
//	 Gain control of target creature until end of turn. Untap that creature.
//	 It gains haste until end of turn."
//
// The flash is a timing permission read as the cast begins (CR 601.3,
// Spec.FlashIf): a crime committed earlier this turn. Casting this very spell
// at an opponent's creature is a crime, but only once it is cast, so it does
// not give itself flash. The effect is Act of Treason's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f815ca4-0197-48c7-bfa2-5eb937c615e6",
		Name:         "Take for a Ride",
		Completeness: CompletenessFull,
		FlashIf: FlashIfYou("as long as you've committed a crime this turn",
			func(g *game.Game, you uuid.UUID) bool { return YouCommittedACrimeThisTurn(g, you) }),
		Targets: TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || !ctx.IsTargetLegal(t) {
				return nil
			}
			return stdB31Threaten(ctx, t, "Take for a Ride")
		},
	})
}
