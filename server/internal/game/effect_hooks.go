package game

import "github.com/google/uuid"

// effect_hooks.go holds the function-variable slots that the S14
// card-effect catalog populates from its own init() block. The
// game package can't directly import server/internal/cards/effects
// (effects needs to reach into *Game for mutations, so that
// direction is the one that builds cleanly). Dependency inversion
// via exported `var` callbacks breaks the cycle:
//
//   game  ──  EffectResolver / ETBEffectHook / IsCatalogCard  ──┐
//                                                                │ populates at init
//   effects (imports game) ───────────────────────────────────────┘
//
// main.go blank-imports effects so the init fires at server boot.
// With no import, the hooks stay nil and the resolution path falls
// back to today's manual sandbox behaviour — the catalog is opt-in
// at the server build level too.
//
// Added in S14 sub-PR 3.

// EffectResolver is invoked from resolveTopOfStackLocked after the
// target re-check and before zone routing. Implementations look up
// the spell's Scryfall ID in the catalog and run the registered
// OnResolve callback if present. Nil means "no catalog wired" —
// the resolution path skips the call and relies on the manual
// sandbox behaviour.
//
// Implementations MUST NOT take g.mu — the resolution path already
// holds it. Use *ForEffect helpers on *Game.
var EffectResolver func(g *Game, item *StackItem, scryfallID string) error

// ETBEffectHook fires after a permanent crosses into the
// battlefield from any source (land cast, spell resolution,
// MoveCardByID into battlefield). Implementations look up the
// card's Scryfall ID in the catalog and run the registered OnETB
// callback + stamp StartingLoyalty for planeswalkers.
var ETBEffectHook func(g *Game, cardID uuid.UUID, scryfallID string) error

// IsCatalogCard reports whether a Scryfall ID is present in the
// card-effect catalog. Used by the view layer (protocol.CardView)
// to stamp the `auto` bit so the client can render the auto badge.
// Nil is treated as "no catalog wired" → every card looks manual.
var IsCatalogCard func(scryfallID string) bool

// fireEffectResolverLocked invokes the registered EffectResolver
// if non-nil, emits EventEffectError on failure, and swallows the
// error so the resolution path keeps moving. Caller must hold g.mu.
func (g *Game) fireEffectResolverLocked(item *StackItem, scryfallID string, cardID uuid.UUID) {
	if EffectResolver == nil {
		return
	}
	if err := EffectResolver(g, item, scryfallID); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Source:   cardID,
			ErrorMsg: err.Error(),
		})
	}
}

// fireETBHookLocked invokes the registered ETBEffectHook if
// non-nil. Used by every code path that places a card on the
// battlefield — land cast, spell resolution, manual admin move
// into battlefield. Caller must hold g.mu.
func (g *Game) fireETBHookLocked(cardID uuid.UUID, scryfallID string) {
	if ETBEffectHook == nil {
		return
	}
	if err := ETBEffectHook(g, cardID, scryfallID); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Source:   cardID,
			ErrorMsg: err.Error(),
		})
	}
}

// isCatalogCardLocked is the internal accessor that the view layer
// uses (via protocol.ViewOfCard) to stamp CardView.Auto. Nil hook
// returns false — no catalog means no auto. Lock-free.
func isCatalogCard(scryfallID string) bool {
	if IsCatalogCard == nil {
		return false
	}
	return IsCatalogCard(scryfallID)
}

// IsAutoCard is the exported wrapper for the view layer. Returns
// true when the card is in the catalog. Called from
// protocol.viewOfCard when stamping CardView.Auto.
func IsAutoCard(scryfallID string) bool {
	return isCatalogCard(scryfallID)
}
