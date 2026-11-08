package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Push the Limit — Sorcery {5}{R}{R} (Aetherdrift, #4464):
//
//	"Return all Mount and Vehicle cards from your graveyard to the
//	 battlefield. Sacrifice them at the beginning of the next end step.
//	 Vehicles you control become artifact creatures until end of turn.
//	 Creatures you control gain haste until end of turn."
//
// The cards are read before anything moves and return together, as
// Splendid Reclamation's lands do (ReturnFromGraveyardTogether), so each
// sees the others enter. The rest of the spell is the return's keyed
// tail, told which permanents really arrived: a delayed trigger
// sacrifices exactly those at the beginning of the next end step,
// whoever's turn it is (encore's sacrificeListedCardsBody; one that has
// left or changed hands is not sacrificed). Then every Vehicle you
// control, the returned ones and any already there, becomes an artifact
// creature until end of turn (The Thanos-Copter's loop), and every
// creature you control at that moment, the freshly animated Vehicles
// included, gains haste until end of turn. Both sets are fixed as the
// spell resolves: a Vehicle or creature that arrives later in the turn
// is not one of them. "Mount" and "Vehicle" are words of the spell's
// text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "62e54c6d-d775-4597-bd4d-dc38460b33ec",
		Name:         "Push the Limit",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			mount, vehicle := ctx.TextSubtype("Mount"), ctx.TextSubtype("Vehicle")
			return ReturnFromGraveyardTogether{
				Targets: graveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
					return c.HasSubtype(mount) || c.HasSubtype(vehicle)
				}),
				ThenTail: pushTheLimitRest,
			}.Apply(ctx)
		},
	})
}

// pushTheLimitRest is everything after the return, told what entered. An
// on-disk key: never renamed, never reused.
var pushTheLimitRest = TogetherTail("push-the-limit/sacrifice-animate-haste", func(ctx *Context, entered []uuid.UUID, _ int) error {
	if len(entered) > 0 {
		if err := (ScheduleDelayedTrigger{
			Label: "Push the Limit — sacrifice the returned Mounts and Vehicles",
			Cards: entered,
			Body:  sacrificeListedCardsBody,
		}).Apply(ctx); err != nil {
			return err
		}
	}
	you := ctx.Controller()
	vehicle := ctx.TextSubtype("Vehicle")
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller != you || !c.HasSubtype(vehicle) {
			continue
		}
		if err := (BecomeCreatureUntilEOT{Target: c.InstanceID, Label: "Push the Limit — becomes an artifact creature"}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return GrantKeywordUntilEOT{
		Match:    And(Creature(), YouControl()),
		Keywords: []string{"haste"},
		Label:    "Push the Limit — haste",
	}.Apply(ctx)
})
