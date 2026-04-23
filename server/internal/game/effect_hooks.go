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

// ManaAbilityShape is the minimal mana-ability surface the game
// package consumes. Mirrors effects.ManaAbility but lives in `game`
// to avoid an import cycle (the effects package already imports
// `game`). The S15 dispatcher reads this on every
// activate_mana_ability call to look up the ability's cost shape +
// produced-mana string.
type ManaAbilityShape struct {
	TapCost       bool
	SacrificeCost bool
	Produced      string
	Label         string
}

// CatalogManaAbilities returns the registered mana abilities for
// the given oracle ID (one entry per `effects.Spec.ManaAbilities`
// element), or nil when no catalog entry exists / the entry has no
// mana abilities. The cards/effects package populates this hook at
// init time alongside EffectResolver / ETBEffectHook. Nil hook ⇒
// engine falls back to the synthetic basic-land ability path.
//
// Added in S15 sub-PR 2.
var CatalogManaAbilities func(oracleID string) []ManaAbilityShape

// CatalogStaticAbilities returns the registered static abilities
// for the given oracle ID (one entry per `effects.Spec.Static`
// element), or nil when no catalog entry exists / the entry has no
// statics. The cards/effects package populates this hook at init
// time alongside the other catalog hooks. Nil hook ⇒ no card has a
// static ability ⇒ the layer engine's recompute pass leaves
// effective characteristics equal to printed.
//
// Used by Game.activeStaticAbilitiesLocked at every recompute pass
// (snapshot-driven, lazy via the layerVersion counter). Added in
// S16 sub-PR 3.
var CatalogStaticAbilities func(oracleID string) []StaticAbility

// CatalogReplacements returns the registered replacement effects
// for the given oracle ID (one entry per
// `effects.Spec.Replacements` element), or nil when no catalog
// entry exists / the entry has no replacements. The cards/effects
// package populates this hook at init time alongside the other
// catalog hooks. Nil hook ⇒ no card has a replacement effect ⇒
// gatherActiveReplacementsLocked skips the catalog leg and returns
// only built-ins + test replacements.
//
// Used by Game.gatherActiveReplacementsLocked at every apply-loop
// iteration. Unlike static abilities, replacement effects fire
// PRE-event — the engine walks the battlefield to find applicable
// effects BEFORE any rule-visible mutation runs. See
// replacements.go. Added in S17 sub-PR 2.
var CatalogReplacements func(oracleID string) []ReplacementEffect

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
