package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// wire.go populates the function-variable hooks declared in
// server/internal/game/effect_hooks.go at package init time.
// Importing this package (blank import from main.go is enough)
// switches the game's resolution path from manual-sandbox-only to
// catalog-aware.
//
// The hooks are set unconditionally — absence of a catalog entry
// still matters (Lookup returns false) and is what preserves the
// opt-in invariant for non-catalog cards.
//
// Added in S14 sub-PR 3.

func init() {
	game.EffectResolver = resolveSpell
	game.ETBEffectHook = fireOnETB
	game.IsCatalogCard = Has
	game.CatalogTargetMode = func(oracleID string) string {
		if spec, ok := Lookup(oracleID); ok {
			return spec.TargetMode
		}
		return ""
	}
	game.CatalogManaAbilities = func(oracleID string) []game.ManaAbilityShape {
		spec, ok := Lookup(oracleID)
		if !ok || len(spec.ManaAbilities) == 0 {
			return nil
		}
		out := make([]game.ManaAbilityShape, len(spec.ManaAbilities))
		for i, a := range spec.ManaAbilities {
			out[i] = game.ManaAbilityShape{
				TapCost:       a.Cost.Tap,
				SacrificeCost: a.Cost.Sacrifice,
				Produced:      a.Produced,
				Label:         a.Label,
			}
		}
		return out
	}
	// S16 sub-PR 3: static abilities flow straight through — the
	// catalog declares them as `game.StaticAbility` already, so the
	// hook is a thin lookup.
	//
	// S18 sub-PR 2: extended to synthesize a self-only Layer 6
	// StaticAbility per card with `PrintedKeywords`, so declaring
	// keywords in the dedicated slot lands them in
	// `card.Effective().Abilities` alongside any hand-written
	// statics. The synthesized ability is idempotent (dedupes
	// against already-present strings) and applies only to the
	// source card itself.
	//
	// Lookup-miss / no-static + no-printed-keywords returns nil so
	// the layer engine treats the card as inert.
	game.CatalogStaticAbilities = func(oracleID string) []game.StaticAbility {
		spec, ok := Lookup(oracleID)
		if !ok {
			return nil
		}
		if len(spec.Static) == 0 && len(spec.PrintedKeywords) == 0 {
			return nil
		}
		if len(spec.PrintedKeywords) == 0 {
			return spec.Static
		}
		// Capture the slice by value so the closure sees a stable
		// list even if a future mutation to `spec.PrintedKeywords`
		// happened (in practice Specs are frozen after Register,
		// but the defensive copy is cheap).
		kws := append([]string(nil), spec.PrintedKeywords...)
		synth := game.StaticAbility{
			Layer:     game.Layer6Ability,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, source *game.Card) {
				for _, kw := range kws {
					if !keywordSliceContains(c.Abilities, kw) {
						c.Abilities = append(c.Abilities, kw)
					}
				}
			},
		}
		if len(spec.Static) == 0 {
			return []game.StaticAbility{synth}
		}
		out := make([]game.StaticAbility, 0, len(spec.Static)+1)
		out = append(out, spec.Static...)
		out = append(out, synth)
		return out
	}
	// S17 sub-PR 2: replacement effects also declared directly as
	// `game.ReplacementEffect`, same thin-lookup shape. Lookup-miss
	// / no-replacements returns nil so the gather pass skips the
	// catalog leg for the card.
	game.CatalogReplacements = func(oracleID string) []game.ReplacementEffect {
		spec, ok := Lookup(oracleID)
		if !ok || len(spec.Replacements) == 0 {
			return nil
		}
		return spec.Replacements
	}
	// S18 sub-PR 2: printed combat keywords. Off-battlefield
	// HasKeyword reads from this hook (the layer engine doesn't
	// maintain Effective() outside the battlefield, so a
	// hand-resident Ambush Viper needs this fallback for flash
	// gating). Lookup-miss / no-printed-keywords returns nil.
	game.CatalogPrintedKeywords = func(oracleID string) []string {
		spec, ok := Lookup(oracleID)
		if !ok || len(spec.PrintedKeywords) == 0 {
			return nil
		}
		return spec.PrintedKeywords
	}
	// S19 sub-PR 1: triggered abilities. Same thin-lookup shape as
	// Static / Replacements. The harvester (game.triggerHarvester)
	// walks the battlefield on every emit and reads this hook per
	// card. Lookup-miss / no-triggers returns nil so the harvester's
	// inner loop continues without allocating.
	game.CatalogTriggers = func(oracleID string) []game.TriggeredAbility {
		spec, ok := Lookup(oracleID)
		if !ok || len(spec.Triggered) == 0 {
			return nil
		}
		return spec.Triggered
	}
}

// selfOnly is the AppliesTo predicate for the synthesized
// printed-keyword StaticAbility — the keyword list applies only to
// the card itself, not to any other battlefield card. Separated
// here so the closure inside the hook is a bare reference rather
// than an inline func literal per call.
func selfOnly(target *game.Card, g *game.Game, source *game.Card) bool {
	return target.InstanceID == source.InstanceID
}

// keywordSliceContains is a small helper for the synthesized
// keyword apply — avoids re-appending a keyword already in the
// slice (e.g. Lord of Atlantis grants "islandwalk" to a Merfolk
// that already has printed islandwalk from its own
// `PrintedKeywords`). Named to avoid colliding with the
// `containsString` helper in test files in this package.
func keywordSliceContains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

// resolveSpell is the EffectResolver implementation. Called from
// resolveTopOfStackLocked after target re-check, before zone
// routing. Looks up the spell's Scryfall ID; runs OnResolve if
// the spec has one. A miss (non-catalog card) or a Spec with nil
// OnResolve (vanilla permanent in the catalog — Birds of Paradise,
// Sol Ring) is a no-op. The resolution path carries on with
// standard zone routing either way.
func resolveSpell(g *game.Game, item *game.StackItem, oracleID string) error {
	spec, ok := Lookup(oracleID)
	if !ok {
		return nil
	}
	if spec.OnResolve == nil {
		return nil
	}
	ctx := NewContext(g, item)
	return spec.OnResolve(item, ctx)
}

// fireOnETB is the ETBEffectHook implementation. Looks up the
// card's Scryfall ID, stamps StartingLoyalty (planeswalkers), and
// runs OnETB if the spec has one. Invoked from every battlefield-
// entry site in the game package.
//
// Looks up the live Card via FindCardZoneForEffect so the hook
// reads the current card state (post-ETB zone move) rather than a
// stale copy captured before the move.
func fireOnETB(g *game.Game, cardID uuid.UUID, oracleID string) error {
	spec, ok := Lookup(oracleID)
	if !ok {
		return nil
	}
	// Stamp starting loyalty first — the SBA check that kills
	// 0-loyalty planeswalkers runs on the next priority-grant
	// boundary, so the counters have to be in place before then.
	if spec.StartingLoyalty > 0 {
		if err := g.AddCounterForEffect(cardID, "loyalty", spec.StartingLoyalty); err != nil {
			return err
		}
	}
	if spec.OnETB == nil {
		return nil
	}
	// Find the live card in the battlefield so the hook receives
	// the current state (post-move).
	zone := g.FindCardZoneForEffect(cardID)
	if zone == nil {
		return nil
	}
	var card *game.Card
	for i := range zone.Cards {
		if zone.Cards[i].InstanceID == cardID {
			card = &zone.Cards[i]
			break
		}
	}
	if card == nil {
		return nil
	}
	ctx := NewContext(g, nil)
	return spec.OnETB(card, ctx)
}
