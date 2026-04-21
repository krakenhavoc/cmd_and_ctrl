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
// the spell's oracle ID (stable across printings) in the catalog
// and run the registered OnResolve callback if present. Nil means
// "no catalog wired" — the resolution path skips the call and
// relies on the manual sandbox behaviour.
//
// Implementations MUST NOT take g.mu — the resolution path already
// holds it. Use *ForEffect helpers on *Game.
var EffectResolver func(g *Game, item *StackItem, oracleID string) error

// ETBEffectHook fires after a permanent crosses into the
// battlefield from any source (land cast, spell resolution,
// MoveCardByID into battlefield). Implementations look up the
// card's oracle ID in the catalog and run the registered OnETB
// callback + stamp StartingLoyalty for planeswalkers.
var ETBEffectHook func(g *Game, cardID uuid.UUID, oracleID string) error

// IsCatalogCard reports whether an oracle ID is present in the
// card-effect catalog. Used by the view layer (protocol.CardView)
// to stamp the `auto` bit so the client can render the auto badge.
// Nil is treated as "no catalog wired" → every card looks manual.
var IsCatalogCard func(oracleID string) bool

// CatalogTargetMode returns the registered card's announce-time
// target prompt shape (see effects.Spec.TargetMode) or empty
// string when no catalog entry matches. Nil hook always returns
// empty. Serialised onto CardView.TargetMode for the client's
// cast-targeting UI.
var CatalogTargetMode func(oracleID string) string

// fireEffectResolverLocked invokes the registered EffectResolver
// if non-nil, emits EventEffectError on failure, and swallows the
// error so the resolution path keeps moving. Caller must hold g.mu.
func (g *Game) fireEffectResolverLocked(item *StackItem, oracleID string, cardID uuid.UUID) {
	if EffectResolver == nil || oracleID == "" {
		return
	}
	if err := EffectResolver(g, item, oracleID); err != nil {
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
func (g *Game) fireETBHookLocked(cardID uuid.UUID, oracleID string) {
	if ETBEffectHook == nil || oracleID == "" {
		return
	}
	if err := ETBEffectHook(g, cardID, oracleID); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Source:   cardID,
			ErrorMsg: err.Error(),
		})
	}
}

// IsAutoCard is the exported wrapper for the view layer. Returns
// true when the card's oracle ID is in the catalog. Called from
// protocol.viewOfCard when stamping CardView.Auto. Nil hook returns
// false — no catalog means no auto.
func IsAutoCard(oracleID string) bool {
	if IsCatalogCard == nil || oracleID == "" {
		return false
	}
	return IsCatalogCard(oracleID)
}

// TargetModeFor returns the catalog's declared target prompt mode
// for the given oracle ID, or empty string if the card isn't in
// the catalog / has no target prompt. Called from protocol.viewOfCard
// when stamping CardView.TargetMode.
func TargetModeFor(oracleID string) string {
	if CatalogTargetMode == nil || oracleID == "" {
		return ""
	}
	return CatalogTargetMode(oracleID)
}
