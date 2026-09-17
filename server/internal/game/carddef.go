package game

import "github.com/google/uuid"

// carddef.go — the one place the engine reads the card catalog
// (#622, Discussion #561 option A).
//
// The game package cannot import the catalog (the catalog imports
// game), so the boundary is a function variable filled in at boot.
// Until #622 there were twenty-two of them, one per Spec slot, each
// doing its own registry lookup — a 368-byte Spec copied by value —
// and several re-projecting slices on every call: the printed-keyword
// static was synthesised, closure and all, on every layer recompute
// for every keyword card. Now there is one lookup that returns a
// *CardDef built once at effects.Register, with every slot already in
// the shape the engine wants.
//
// The per-slot variables (CatalogTriggers, CatalogStaticAbilities,
// EffectResolver, …) still exist, because two dozen test files stub
// them individually to inject catalog behaviour without importing the
// catalog. They are no longer set by the catalog: init below gives
// each a default that reads its slot off the CardDef, so production
// has one hook and a test can still override one slot. Adding a slot
// is a field on effects.Spec, a field here, one line in
// effects.buildDef, and the call site — plus a per-slot variable only
// if a test needs to stub it.

// CardDef is a catalog card as the engine reads it: every Spec slot,
// precomputed. Built once per card; never mutated after boot, so the
// slices it hands out are shared and read-only.
type CardDef struct {
	// Resolve runs the card's OnResolve; nil when it has none.
	Resolve func(g *Game, item *StackItem) error
	// AsEnters runs the card's CR 614.12 "as enters" hook; nil for
	// nearly every card.
	AsEnters func(g *Game, cardID uuid.UUID) error

	StartingLoyalty int
	BattleDefense   int

	// TargetMode is the client hint: the structured clause's Mode
	// when there is one, else the legacy free-form string.
	TargetMode string
	Targets    *TargetSpec
	Modes      *ModeSpec

	ManaAbilities []ManaAbilityShape
	Activated     []ActivatedAbilityShape
	// Static includes the Layer 6 keyword static synthesised from
	// PrintedKeywords, appended once at build time.
	Static          []StaticAbility
	Replacements    []ReplacementEffect
	PrintedKeywords []string
	Triggered       []TriggeredAbility
	TriggerDoublers []TriggerDoubler

	AdditionalCost   *AdditionalCost
	AlternativeCosts []AlternativeCost
	TapCost          *TapPermanentsCost
	CostModifiers    []CostModifier
	// SelfCostModifiers change what THIS card costs to cast (ADR 0048
	// addendum §11), read by SelfCostModifiersFor for the spell being
	// priced and never from the battlefield.
	SelfCostModifiers     []CostModifier
	CastableZones         []ZoneKind
	UntapStep             []UntapStepPermission
	UntapStepRestrictions []UntapStepRestriction

	CantBeCountered bool
	NoMaxHandSize   bool
}

// CatalogLookup is the one production hook: the catalog's definition
// for a CatalogKey, or nil when the card has no entry. Set by the
// effects package at boot; nil means no catalog is wired and every
// card is a manual sandbox card.
var CatalogLookup func(key string) *CardDef

// catalogDef is CatalogLookup with the nil-hook and empty-key guards
// every caller wants.
func catalogDef(key string) *CardDef {
	if CatalogLookup == nil || key == "" {
		return nil
	}
	return CatalogLookup(key)
}

func init() {
	EffectResolver = func(g *Game, item *StackItem, key string) error {
		if d := catalogDef(key); d != nil && d.Resolve != nil {
			return d.Resolve(g, item)
		}
		return nil
	}
	ETBEffectHook = func(g *Game, cardID uuid.UUID, key string) error {
		if d := catalogDef(key); d != nil && d.AsEnters != nil {
			return d.AsEnters(g, cardID)
		}
		return nil
	}
	IsCatalogCard = func(key string) bool { return catalogDef(key) != nil }
	CatalogStartingLoyalty = func(key string) int {
		if d := catalogDef(key); d != nil {
			return d.StartingLoyalty
		}
		return 0
	}
	CatalogBattleDefense = func(key string) int {
		if d := catalogDef(key); d != nil {
			return d.BattleDefense
		}
		return 0
	}
	CatalogTargetMode = func(key string) string {
		if d := catalogDef(key); d != nil {
			return d.TargetMode
		}
		return ""
	}
	CatalogTargetSpec = func(key string) *TargetSpec {
		if d := catalogDef(key); d != nil {
			return d.Targets
		}
		return nil
	}
	CatalogModeSpec = func(key string) *ModeSpec {
		if d := catalogDef(key); d != nil {
			return d.Modes
		}
		return nil
	}
	CatalogManaAbilities = func(key string) []ManaAbilityShape {
		if d := catalogDef(key); d != nil {
			return d.ManaAbilities
		}
		return nil
	}
	CatalogActivatedAbilities = func(key string) []ActivatedAbilityShape {
		if d := catalogDef(key); d != nil {
			return d.Activated
		}
		return nil
	}
	CatalogStaticAbilities = func(key string) []StaticAbility {
		if d := catalogDef(key); d != nil {
			return d.Static
		}
		return nil
	}
	CatalogReplacements = func(key string) []ReplacementEffect {
		if d := catalogDef(key); d != nil {
			return d.Replacements
		}
		return nil
	}
	CatalogPrintedKeywords = func(key string) []string {
		if d := catalogDef(key); d != nil {
			return d.PrintedKeywords
		}
		return nil
	}
	CatalogTriggers = func(key string) []TriggeredAbility {
		if d := catalogDef(key); d != nil {
			return d.Triggered
		}
		return nil
	}
	CatalogTriggerDoublers = func(key string) []TriggerDoubler {
		if d := catalogDef(key); d != nil {
			return d.TriggerDoublers
		}
		return nil
	}
	CatalogAdditionalCost = func(key string) *AdditionalCost {
		if d := catalogDef(key); d != nil {
			return d.AdditionalCost
		}
		return nil
	}
	CatalogAlternativeCosts = func(key string) []AlternativeCost {
		if d := catalogDef(key); d != nil {
			return d.AlternativeCosts
		}
		return nil
	}
	CatalogTapPermanentsCost = func(key string) *TapPermanentsCost {
		if d := catalogDef(key); d != nil {
			return d.TapCost
		}
		return nil
	}
	CatalogCostModifiers = func(key string) []CostModifier {
		if d := catalogDef(key); d != nil {
			return d.CostModifiers
		}
		return nil
	}
	CatalogCastableZones = func(key string) []ZoneKind {
		if d := catalogDef(key); d != nil {
			return d.CastableZones
		}
		return nil
	}
	CatalogUntapStepPermissions = func(key string) []UntapStepPermission {
		if d := catalogDef(key); d != nil {
			return d.UntapStep
		}
		return nil
	}
	CatalogUntapStepRestrictions = func(key string) []UntapStepRestriction {
		if d := catalogDef(key); d != nil {
			return d.UntapStepRestrictions
		}
		return nil
	}
	CatalogCantBeCountered = func(key string) bool {
		d := catalogDef(key)
		return d != nil && d.CantBeCountered
	}
	CatalogNoMaxHandSize = func(key string) bool {
		d := catalogDef(key)
		return d != nil && d.NoMaxHandSize
	}
}
