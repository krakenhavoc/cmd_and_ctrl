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
//   - **No foretell** (CR 702.143, tracked in #658). Foretell is three
//     things the engine does not have yet, and none of them is a
//     card-file detail. First, a special action taken from HAND: pay
//     {2} and exile the card face down, any time its owner has
//     priority during their own turn (CR 116.2h, 702.143a-b); there
//     is no special-action verb (#655). Second, a face-down exile its
//     OWNER can see and the table cannot (CR 702.143a): both
//     primitives that mint an exile-play grant
//     (ExileTopWithPermissionForEffect,
//     ExileCardWithPermissionForEffect) mark the card known to every
//     seat on the way in, deliberately, because an impulse grant
//     nobody can see is unplayable, and the face-down route marks it
//     known to nobody (#656). Third, a permission on that ONE exiled
//     card, keyed "foretell" and live from the next turn on, that
//     leaves the spell "foretold" on the stack (CR 702.143c-d, #652).
//     `Spec.CastableZones` plus `AlternativeCost.FromZone` are NOT
//     that permission: they are card-level, so they would make every
//     exiled copy, a Path to Exile target included, castable for
//     {1}{W} on any turn. Warp's end-step exile
//     (scheduleWarpExileLocked) is the nearest precedent for the
//     third piece. The card is castable only for its printed {3}{W}
//     until foretell lands. Strictly weaker than printed.
//   - The replacement does not fire on a permanent that would go to
//     the **command zone** instead (a commander dying with the CR
//     903.9 built-in taken): that built-in rewrites the destination
//     first, and by the time this effect sees the event the move is
//     no longer battlefield → graveyard. Paper resolves the two as a
//     choice between simultaneous replacements; here the commander
//     one wins.
func init() {
	Register(Spec{
		OracleID:     "cddccc2a-a76e-48b3-b4dd-dfeab89e1619",
		Name:         "Cosmic Intervention",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Foretell isn't implemented, so it can only be cast for its normal cost; a dying commander still goes to the command zone instead of being saved."},
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
