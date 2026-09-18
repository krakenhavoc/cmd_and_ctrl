package effects

import (
	"errors"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Path to Exile — "Exile target creature. Its controller may
// search their library for a basic land card, put it onto the
// battlefield tapped, then shuffle."
//
// Capture the controller BEFORE ExileTarget fires — post-move the
// card is in exile and its controller is still stamped, but the
// lookup is cleaner to do upfront.
//
// S22: the "may" is real. The search is Optional, so the creature's
// controller is prompted and can decline the land — and with it the
// shuffle, which is the half a player with a stacked top of library
// actually cares about. S14 treated may as always.
//
// S17 sub-PR 4: fetched land enters TAPPED via
// SearchLibrary.TappedOnEntry. Closes the S14 "enters untapped"
// deferral; matches the card text.
//
// #894: the search runs from the exile's continuation, so the table
// answers the CR 903.9 command-zone question first and is offered the
// basic-land search afterwards.
func init() {
	Register(Spec{
		OracleID:     "d683d985-9888-4d21-8b5f-69e69ce4a03b",
		Name:         "Path to Exile",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
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
			return ExileTarget{
				Target: targetID,
				// #894: the search is the exile's continuation, not the
				// next line. Exiling a commander opens the CR 903.9
				// window, and the search used to be offered while that
				// question was still on the table — two prompts at
				// once, in the wrong order. The `exiled` answer is
				// deliberately ignored: "its controller may search" is
				// a separate sentence rather than an "if you do", so
				// the search happens either way and only its ORDER
				// changes.
				Then: func(ctx *Context, _ bool) error {
					// Search the controller's library for any basic land.
					return SearchLibrary{
						Player:        controller,
						Predicate:     IsBasicLand,
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Reveal:        true,
						Shuffle:       true,
						TappedOnEntry: true,
						Optional:      true,
						Reason:        "Path to Exile — you may search for a basic land",
					}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}
