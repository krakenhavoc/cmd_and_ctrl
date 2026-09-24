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
// History — both S22 simplifications are closed:
//
//   - **Foretell** (CR 702.143) ships on #658 and is one line on this
//     card: `Foretell("{1}{W}")`. The comment this replaced claimed
//     `Spec.CastableZones` plus `AlternativeCost.FromZone` already
//     covered the cost half and only the hidden exile was missing.
//     That was wrong twice. Those two are CARD-level declarations, so
//     they would make every exiled copy of this card — a Path to
//     Exile target included — castable for {1}{W} on any turn, while
//     foretold status and the foretell cost belong to the ONE card
//     that was foretold (CR 702.143c-d). And they are not "the cost
//     half" either: the foretell cost is claimable only from a LATER
//     turn (CR 702.143a), which no card-level declaration can say.
//     What foretell actually rides is a per-instance CastPermission
//     (ADR 0066) over that one exiled object, carrying the
//     NotBeforeSeq floor, plus ADR 0069's face-down kind for the
//     owner-only look.
//
// A dying COMMANDER is saved too, and its owner decides. The old
// caveat here ("a dying commander still goes to the command zone")
// described an engine that no longer exists. CR 903.9a (pinned
// edition) makes a commander's trip from a graveyard or exile a
// state-based "may" AFTER it arrives, so in paper this replacement
// exiles the commander and its owner then chooses: command zone now,
// or stay exiled and come back at the end step. The engine models the
// command-zone rule as an optional replacement on the same move
// (builtin_replacements.go), and the two meet in the CR 616 apply
// loop: the controller orders them, the commander offer is asked, and
// "no" leaves this effect to exile the card and schedule its return.
// Both printed outcomes are reachable and nothing else is — pinned by
// TestCosmicInterventionSavesACommanderWhoseOwnerDeclines.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cddccc2a-a76e-48b3-b4dd-dfeab89e1619",
		Name:         "Cosmic Intervention",
		Completeness: CompletenessFull,
		SpecialActions: []game.SpecialAction{
			Foretell("{1}{W}"),
		},
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
						Body:         returnExiledToOwnersBody.Key(),
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
