package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shuffle_into_library_instead.go — "If this card would be put into a
// graveyard from anywhere, reveal it and shuffle it into its owner's
// library instead" (Blightsteel Colossus, Nexus of Fate).
//
// The self-replacement is a `RepEventMove` whose destination is
// rewritten before the move happens, with no `OldZone` check at all:
// that omission IS "from anywhere" (#539 made every zone exit open this
// same window, not only a battlefield death, and the stack's own
// resolution and a counterspell go through it too).
//
// "From anywhere" also means a DISCARD, and a discard is its own
// ReplacementEventKind (RepEventDiscard, not RepEventMove — ADR 0013
// §5g), pre-filtered by its own Watches key (EventDiscardCard, not
// EventZoneMove). A card that watched only the move kind would let a
// discarded card fall straight into the graveyard, so both are
// declared and AppliesTo accepts either Kind: mill and a battlefield
// death already carry RepEventMove, and discard carries RepEventDiscard
// with the same NewZone/CardID shape (the two "share every field the
// exit path reads", per replacements.go's isExitMove).
//
// The card is genuinely shuffled into its owner's library, not merely
// placed on top (ADR 0013 §5ah's ShuffleDestinationLibrary): the
// printed card reveals WHICH card, never WHERE it ends up.

// ShuffleIntoOwnersLibraryInstead is the self-replacement for the
// clause above. `label` is the replacement's prompt and log name,
// "<Card>: shuffled into its owner's library instead".
func ShuffleIntoOwnersLibraryInstead(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:         []game.EventKind{game.EventZoneMove, game.EventDiscardCard},
		SelfReplacement: true,
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			fromMove := ev.Kind == game.RepEventMove || ev.Kind == game.RepEventDiscard
			return fromMove && ev.CardID == src.InstanceID && ev.NewZone == game.ZoneGraveyard
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.NewZone = game.ZoneLibrary
			ev.NewZoneOwner = uuid.Nil
			ev.ShuffleDestinationLibrary = true
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}
