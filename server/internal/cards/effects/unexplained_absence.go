package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unexplained Absence — Instant {3}{W} (Murders at Karlov Manor Commander,
// #5368):
//
//	"For each player, exile up to one target nonland permanent that player
//	 controls. For each permanent exiled this way, its controller cloaks
//	 the top card of their library."
//
// One "up to one" clause per seat, counted from the caster (the table has
// at most four). The exile is one simultaneous move; the batch's
// last-known copies say who controlled each permanent, and each of them
// cloaks once per permanent they lost (CR 701.58a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "71faaee5-70ff-449c-85c8-560eb189a0a7",
		Name:         "Unexplained Absence",
		Completeness: CompletenessFull,
		Targets:      cmdr2S38SeatClauses(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			if len(ids) == 0 {
				return nil
			}
			return ctx.Game.ExileCardsKeyedForEffect(ids, ctx.Tail(unexplainedAbsenceCloak))
		},
	})
}

var unexplainedAbsenceCloak = MassTail("mass/unexplained-absence-cloak", func(ctx *Context, swept []game.Card, _ int) error {
	for _, c := range swept {
		if _, err := (CloakTopOfLibrary{Player: c.Controller}).Do(ctx); err != nil {
			return err
		}
	}
	return nil
})
