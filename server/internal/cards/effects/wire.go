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
	// #274: starting loyalty is printed card data, so the stamp
	// itself lives in the game package. The catalog only answers
	// when the card carries no printed value of its own.
	game.CatalogStartingLoyalty = func(oracleID string) int {
		if spec, ok := Lookup(oracleID); ok {
			return spec.StartingLoyalty
		}
		return 0
	}
	// S27: a battle's printed defense. Same fallback-only contract
	// the starting-loyalty hook has — the catalog answers only when
	// the card carries no printed value of its own.
	game.CatalogBattleDefense = func(oracleID string) int {
		if spec, ok := Lookup(oracleID); ok && spec.Battle != nil {
			return spec.Battle.Defense
		}
		return 0
	}
	game.CatalogTargetMode = func(oracleID string) string {
		if spec, ok := Lookup(oracleID); ok {
			// S20: a structured TargetSpec is the source of truth
			// for the client hint too.
			if spec.Targets != nil {
				return spec.Targets.Mode
			}
			return spec.TargetMode
		}
		return ""
	}
	// S20 sub-PR 1: structured targeting. Nil for cards without a
	// TargetSpec — the engine falls back to the free-form picker.
	game.CatalogTargetSpec = func(oracleID string) *game.TargetSpec {
		if spec, ok := Lookup(oracleID); ok {
			return spec.Targets
		}
		return nil
	}
	// S21 sub-PR 2: activated abilities. Nil for cards with none.
	game.CatalogActivatedAbilities = func(oracleID string) []game.ActivatedAbilityShape {
		spec, ok := Lookup(oracleID)
		if !ok || len(spec.Activated) == 0 {
			return nil
		}
		out := make([]game.ActivatedAbilityShape, len(spec.Activated))
		for i, a := range spec.Activated {
			out[i] = game.ActivatedAbilityShape{
				Label:        a.Label,
				Cost:         a.Cost,
				Targets:      a.Targets,
				SorcerySpeed: a.SorcerySpeed,
				Effect:       a.Effect,
			}
		}
		return out
	}
	// S20 sub-PR 4: modal spells. Nil for non-modal cards.
	game.CatalogModeSpec = func(oracleID string) *game.ModeSpec {
		if spec, ok := Lookup(oracleID); ok {
			return spec.Modes
		}
		return nil
	}
	// S21 sub-PR 5: additional costs on cast. Nil for cards without
	// one, which is nearly all of them.
	game.CatalogAdditionalCost = func(oracleID string) *game.AdditionalCost {
		if spec, ok := Lookup(oracleID); ok {
			return spec.AdditionalCost
		}
		return nil
	}
	// S22: alternative costs to cast — overload, evoke, cleave. Nil
	// for cards that offer none, which is nearly all of them.
	game.CatalogAlternativeCosts = func(oracleID string) []game.AlternativeCost {
		spec, ok := Lookup(oracleID)
		if !ok || len(spec.AlternativeCosts) == 0 {
			return nil
		}
		return spec.AlternativeCosts
	}
	// S22: convoke / waterbend — tapping permanents to help pay.
	// Nil for cards that offer none, which is nearly all of them.
	game.CatalogTapPermanentsCost = func(oracleID string) *game.TapPermanentsCost {
		if spec, ok := Lookup(oracleID); ok {
			return spec.TapCost
		}
		return nil
	}
	game.CatalogManaAbilities = func(oracleID string) []game.ManaAbilityShape {
		spec, ok := Lookup(oracleID)
		if !ok || len(spec.ManaAbilities) == 0 {
			return nil
		}
		out := make([]game.ManaAbilityShape, len(spec.ManaAbilities))
		for i, a := range spec.ManaAbilities {
			out[i] = game.ManaAbilityShape{
				TapCost:        a.Cost.Tap,
				SacrificeCost:  a.Cost.Sacrifice,
				SacrificeOther: a.Cost.SacrificeOther,
				LifeCost:       a.Cost.Life,
				Produced:       a.Produced,
				Label:          a.Label,
				// S22 mana-ability riders: the post-production
				// callback and the commander-identity opt-out both
				// flow straight through, same thin projection the
				// cost fields get.
				Rider:                   a.Rider,
				IgnoreCommanderIdentity: a.IgnoreCommanderIdentity,
				// S32 mana pipeline (#352): the mana cost
				// component, the activation gate, the computed
				// produced string and the spend restrictions ride
				// the same thin projection everything else does.
				ManaCost:     a.Cost.Mana,
				Condition:    a.Condition,
				ProducedFunc: a.ProducedFunc,
				Restrictions: a.Restrictions,
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
	// #338: the one player-scoped continuous effect in the catalog.
	// Derived on demand at cleanup rather than written into
	// Player.MaxHandSize — see Spec.NoMaxHandSize.
	game.CatalogNoMaxHandSize = func(oracleID string) bool {
		spec, ok := Lookup(oracleID)
		return ok && spec.NoMaxHandSize
	}
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
	// Starting loyalty is NOT stamped here any more. It moved to
	// game.stampStartingLoyaltyLocked, which the game package runs
	// on every battlefield entry whether or not a catalog entry
	// exists — see issue #274. The catalog's Spec.StartingLoyalty
	// is still consulted, as a fallback, via the
	// game.CatalogStartingLoyalty hook registered in init().
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
