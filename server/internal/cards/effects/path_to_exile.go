package effects

import (
	"errors"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Path to Exile — "Exile target creature. Its controller may
// search their library for a basic land card, put it onto the
// battlefield tapped, then shuffle."
//
// S14 sandbox simplifications retained:
//   - "May" is treated as always. If the exiled creature's
//     controller has a basic land, it always fetches.
//   - Capture the controller BEFORE ExileTarget fires — post-move
//     the card is in exile and its controller is still stamped,
//     but the lookup is cleaner to do upfront.
//
// S17 sub-PR 4: fetched land enters TAPPED via
// SearchLibrary.TappedOnEntry. Closes the S14 "enters untapped"
// deferral; matches the card text.
func init() {
	Register(Spec{
		OracleID:   "d683d985-9888-4d21-8b5f-69e69ce4a03b",
		Name:       "Path to Exile",
		TargetMode: "creature",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			// Diagnostic: surface silent no-ops so they're visible in
			// the event log as EventEffectError. Path appearing to
			// "have no effect" was a user-report regression during
			// S17 manual testing — without these returns the spell
			// resolved to graveyard without any trace.
			if len(item.Targets) == 0 {
				return errors.New("Path to Exile: cast with no target (client UI skipped targeting?)")
			}
			if item.Targets[0].Kind != game.TargetCard {
				return errors.New("Path to Exile: first target is not a card")
			}
			targetID := item.Targets[0].ID
			card, ok := ctx.Game.LookupCardForEffect(targetID)
			if !ok {
				return errors.New("Path to Exile: target card not found in any zone")
			}
			controller := card.Controller
			if err := (ExileTarget{Target: targetID}).Apply(ctx); err != nil {
				return err
			}
			// Search the controller's library for any basic land.
			return SearchLibrary{
				Player:        controller,
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Reveal:        true,
				Shuffle:       true,
				TappedOnEntry: true,
			}.Apply(ctx)
		},
	})
}
