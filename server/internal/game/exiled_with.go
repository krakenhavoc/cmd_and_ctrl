package game

import "github.com/google/uuid"

// exiled_with.go — CR 607.2a's "exiled with [this permanent]" for a
// card a REPLACEMENT EFFECT exiled (#2530, ADR 0066's 2026-10-07
// amendment).
//
// Valgavoth, Terror Eater is the first card: "if a card you didn't
// control would be put into an opponent's graveyard from anywhere,
// exile it instead. During your turn, you may play cards exiled with
// Valgavoth." The exile is a replacement, which runs before the card
// moves, so there is no moment inside it at which a card in exile
// exists to be marked. The link therefore travels on the event
// (ReplacementEvent.ExiledWith) and lands with the card, here.
//
// It is a card-carried link rather than the effects package's event-log
// record (b27ExiledWith) because that record is keyed on an ABILITY
// resolving, and a replacement is not one: nothing is resolving, there
// is no EventResolve to open a window on. It is also read by a pure-data
// PermissionFilter, which has a Card in hand and no log to walk.

// stampExiledWithLocked writes the "exiled with" link onto the card
// that has just landed in `dst`, when `dst` is exile and the move
// declared a link. A no-op for every other destination, for the zero
// ref, and for a card that is not in `dst` (a failed move).
//
// Called after MoveCard, which has already cleared any older link:
// the stamp names the move that just happened and no earlier one.
func stampExiledWithLocked(dst *Zone, cardID uuid.UUID, ref PermissionCardRef) {
	if dst == nil || dst.Kind != ZoneExile || ref.ID == uuid.Nil {
		return
	}
	for i := range dst.Cards {
		if dst.Cards[i].InstanceID == cardID {
			dst.Cards[i].ExiledWith = ref
			return
		}
	}
}
