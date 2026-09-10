package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cosmic Intervention — "If a permanent you control would be put
// into a graveyard from the battlefield this turn, exile it instead.
// Return it to the battlefield under its owner's control at the
// beginning of the next end step. // Foretell {1}{W}"
//
// A board wipe answered by a CR 614 replacement plus a CR 603.7
// delayed trigger: the wipe still resolves, but every permanent it
// would have killed lands in exile and walks back at the end step,
// as a new object. Registered as a turn-scoped replacement in
// OnResolve for the Fog reason — the instant is in the graveyard by
// the time the replacement has anything to do, so the effect has to
// outlive its source. StepCleanup clears it.
//
// It catches sacrifice as well as destruction, which is correct:
// the card says "put into a graveyard from the battlefield", not
// "destroyed", and `sacrificePermanentLocked` routes through the
// same replacement pipeline.
//
// One delayed trigger is scheduled per exiled permanent rather than
// one carrying the whole list, because the replacement fires once
// per move event — a wrath produces one event per creature. The
// observable difference is only how many items the stack shows.
//
// S22 sandbox simplifications:
//
//   - **No foretell.** There is no alternative-cast-cost machinery
//     ("pay X *instead of* the mana cost"); `AdditionalCost` is an
//     extra cost paid alongside. The card is castable only for its
//     printed {3}{W}, exactly as Vandalblast and Cyclonic Rift ship
//     without overload. Strictly weaker than printed.
//   - The replacement does not fire on a permanent that would go to
//     the **command zone** instead (a commander dying with the CR
//     903.9 built-in taken): that built-in rewrites the destination
//     first, and by the time this effect sees the event the move is
//     no longer battlefield → graveyard. Paper resolves the two as a
//     choice between simultaneous replacements; here the commander
//     one wins.
func init() {
	Register(Spec{
		OracleID: "cddccc2a-a76e-48b3-b4dd-dfeab89e1619",
		Name:     "Cosmic Intervention",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			// Capture plain IDs, never pointers: the replacement is
			// value-copied onto the undo stack and has to keep
			// working against whichever game it is restored into.
			controller := ctx.Controller()
			source := ctx.Source()
			ctx.Game.RegisterTurnScopedReplacement(game.ReplacementEffect{
				Watches: []game.EventKind{game.EventZoneMove},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
					if ev.Kind != game.RepEventMove {
						return false
					}
					if ev.OldZone != game.ZoneBattlefield || ev.NewZone != game.ZoneGraveyard {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.Controller == controller
				},
				Replace: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) error {
					ev.NewZone = game.ZoneExile
					ev.NewZoneOwner = uuid.Nil
					g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
						Controller:   controller,
						SourceCardID: source,
						Label:        "Cosmic Intervention — return the exiled permanent",
						At:           game.StepEnd,
						Cards:        []uuid.UUID{ev.CardID},
						Effect:       returnExiledCardsToOwners,
					})
					return nil
				},
				Controller: func(_ *game.ReplacementEvent, _ *game.Game, _ *game.Card) uuid.UUID {
					return controller
				},
				Label: "Cosmic Intervention: exile instead of graveyard",
			})
			return nil
		},
	})
}
