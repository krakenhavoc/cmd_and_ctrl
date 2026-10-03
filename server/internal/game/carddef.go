package game

import (
	"strings"

	"github.com/google/uuid"
)

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
	// AsTransformsInto runs the face's "As this permanent transforms
	// into <this face>, …" clause (Sephiroth, One-Winged Angel's Super
	// Nova), registered on the face being turned TO. The transform
	// counterpart of AsEnters: off the stack, run by the in-place verb
	// (TransformPermanentForEffect) whatever effect asked for the
	// transform. Nil for nearly every face. #1574, ADR 0079 amendment
	// 2026-09-24.
	AsTransformsInto func(g *Game, cardID uuid.UUID) error

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
	Static       []StaticAbility
	Replacements []ReplacementEffect
	// EntersWithCountersFromCast are the card's printed "this
	// permanent enters with N counters on it" clauses whose N is read
	// from the spell that became it (CR 614.1c) — Hangarback Walker's
	// X, Etched Oracle's sunburst. Seeded onto the entry event before
	// the CR 614 pipeline; see entry_counters.go.
	EntersWithCountersFromCast []EntryCountersFromCast
	PrintedKeywords            []string
	Triggered                  []TriggeredAbility
	TriggerDoublers            []TriggerDoubler
	// TriggerSuppressors are the "[an event] doesn't cause abilities to
	// trigger" statics (#1735, trigger_suppression.go).
	TriggerSuppressors []TriggerSuppressor
	// ManaTriggers are the CR 605.1b TRIGGERED MANA abilities — the
	// one trigger kind that never uses the stack (CR 605.4a). Kept
	// apart from Triggered because nothing on TriggeredAbility applies
	// to them; see mana_trigger.go and ADR 0074.
	ManaTriggers []ManaTrigger

	AdditionalCost *AdditionalCost
	// OptionalCosts are the additional costs the caster may CHOOSE to
	// pay (ADR 0073) — kicker, multikicker, buyback. The slice order
	// is the index space the announcement and PaidCost.OptionalCosts
	// both name, so it is never re-sorted.
	OptionalCosts    []AdditionalCost
	AlternativeCosts []AlternativeCost
	TapCost          *TapPermanentsCost
	// Delve is CR 702.66: this spell may exile cards from its caster's
	// graveyard to pay generic mana (ADR 0100). Read through DelveFor.
	Delve bool
	// SpellsYouCastHaveDelve is a permanent's "Spells you cast have
	// delve" (Teval, Arbiter of Virtue; ADR 0100 sub-PR 2). Read from
	// the battlefield through DelveForLocked.
	SpellsYouCastHaveDelve bool
	CostModifiers          []CostModifier
	// SelfCostModifiers change what THIS card costs to cast (ADR 0048
	// addendum §11), read by SelfCostModifiersFor for the spell being
	// priced and never from the battlefield.
	SelfCostModifiers []CostModifier
	// ExhaustPermissions are the "you may activate exhaust abilities
	// as though they haven't been activated" statics this permanent
	// contributes (#1184) — Elvish Refueler. Read from the
	// battlefield through ExhaustPermissionsForCard.
	ExhaustPermissions []ExhaustPermission
	// AttackTaxes are the "creatures can't attack you unless their
	// controller pays {N}" statics this permanent contributes
	// (CR 508.1a, ADR 0080) — Propaganda, Ghostly Prison, Windborn
	// Muse. Read from the battlefield through AttackTaxesForCard.
	AttackTaxes []AttackTax
	// BlockRules are the CR 509.1b block restrictions with a
	// parameter this permanent imposes while it is on the battlefield
	// — "can't be blocked except by Walls" (Prowler's Helm), "can't
	// be blocked by more than one creature" (Vorrac Battlehorns). Read
	// through CatalogBlockRules, keyed by CatalogAbilityKey; see
	// block_rules.go and ADR 0045's addendum, Decision 11.
	BlockRules []BlockRule
	// AttackLimits are the CR 508.1c count limits this permanent
	// imposes on an attack declaration — "no more than one creature
	// can attack each combat" (Silent Arbiter), "no more than two
	// creatures can attack you each combat" (Crawlspace). Read through
	// CatalogAttackLimits, keyed by CatalogAbilityKey; see
	// attack_limits.go and ADR 0045 Decision 44 (#1507).
	AttackLimits []AttackLimit
	// HexproofBypasses are this permanent's "can be the targets of
	// spells and abilities as though they didn't have hexproof"
	// statics (CR 702.11) — Nowhere to Run, Kaya, Bane of the Dead.
	// WardSuppressions are its "ward abilities of those creatures
	// don't trigger" statics (CR 702.21). Both read from the
	// battlefield through CatalogAbilityKey; see hexproof_bypass.go
	// and ADR 0038's amendment of 2026-09-27 (#1560).
	HexproofBypasses []HexproofBypass
	WardSuppressions []WardSuppression
	CastableZones    []ZoneKind

	// SpecialActions are the CR 116.2 special actions the card offers
	// from its owner's hand — foretell (CR 702.143a) and suspend
	// (CR 702.62a). Read by PerformSpecialAction and by the
	// legal-move enumerator through SpecialActionsFor. ADR 0062
	// Decision 4.
	SpecialActions []SpecialAction

	// SpecialActionGrants are the special actions this PERMANENT gives
	// to other cards (#1391, Fblthp, Lost on the Range's plot from the
	// top of your library). Read through CatalogSpecialActionGrants;
	// see special_action_grant.go.
	SpecialActionGrants []SpecialActionGrant

	UntapStep             []UntapStepPermission
	UntapStepRestrictions []UntapStepRestriction
	UntapCaps             []UntapCap
	UntapOptOuts          []UntapOptOut

	// DrawStep are the "you draw a card during each opponent's draw
	// step" permissions this source contributes (#1315) —
	// draw_step.go's widening of CR 504.1's turn-based draw, the same
	// shape UntapStep is for CR 502.3's untap. Read from the
	// battlefield AND from every seat's emblem zone through
	// activeDrawStepPermissionsLocked, because CR 114.3 runs an
	// emblem's abilities in the command zone exactly like a
	// permanent's.
	DrawStep []DrawStepPermission

	// CastCondition is the card's own "you may cast this only if …"
	// (CR 205.4e's legendary sorcery, and the "cast only if" family),
	// checked by CastGateLocked at announce and never at resolution.
	// Nil for every card that prints no such clause. ADR 0073 §7.
	CastCondition func(g *Game, controller uuid.UUID, card Card) bool
	// CastConditionLabel is that clause as printed, returned to the
	// client when the gate refuses the cast.
	CastConditionLabel string
	// CastRestrictions are the "can't cast" statics this PERMANENT
	// imposes on other players' casts (Rule of Law, Grafdigger's
	// Cage, Rakdos). Read from the battlefield through
	// CatalogAbilityKey, never from a card's own zone.
	CastRestrictions []CastRestriction
	// LandPlayRestrictions are the "can't play lands" statics this
	// PERMANENT imposes (ADR 0109 §4, #1895). Read from the battlefield
	// through CatalogAbilityKey, like CastRestrictions.
	LandPlayRestrictions []LandPlayRestriction
	// ActivationRestrictions are the "can't be activated" statics
	// this PERMANENT imposes on other objects' activated abilities
	// (Cursed Totem, Linvala, Collector Ouphe, Pithing Needle). Read
	// from the battlefield through CatalogAbilityKey, never from a
	// card's own zone. #1210, ADR 0073's amendment of 2026-09-22.
	ActivationRestrictions []ActivationRestriction
	// ActivationTimings are the per-player activation-TIMING
	// statements this PERMANENT makes while it is on the battlefield
	// (#1208) — The Wandering Emperor's "you may activate her
	// loyalty abilities any time you could cast an instant", Leonin
	// Shikari's "you may activate equip abilities any time you could
	// cast an instant". Read from the battlefield through
	// CatalogAbilityKey; see activation_timing.go.
	//
	// The activation twin of CastTimings below, and it has only this
	// one home: nothing stores an activation-timing statement,
	// because every card that prints one is a permanent that says it
	// for as long as it is there.
	ActivationTimings []ActivationTiming

	CantBeCountered bool
	// CantBeCounteredIf is the CONDITIONAL "this spell can't be
	// countered" (Banefire's X of 5 or more, Demonfire's hellbent),
	// judged when something tries to counter the spell. Nil is never.
	CantBeCounteredIf SpellCondition
	// SpellDamageCantBePrevented is the spell's own "the damage can't
	// be prevented" (CR 615.12, ADR 0107 §5), under its condition. Nil
	// is never; see unpreventable_damage.go.
	SpellDamageCantBePrevented SpellCondition
	// SpellsCantBeCountered are this permanent's printed "<these>
	// spells can't be countered" statics (ADR 0106 §4, #1806) —
	// Chimil, the Inner Sun. Read from the battlefield through
	// CounterShieldsForCard, keyed by CatalogAbilityKey; see
	// cant_be_countered.go.
	SpellsCantBeCountered []CounterShieldStatic
	NoMaxHandSize         bool
	// NoMaxHandSizeWhen gates NoMaxHandSize on a designation (ADR
	// 0071, ADR 0103: Steaming Sauna's door). Zero is no gate.
	NoMaxHandSizeWhen Designation
	// PlayerKeywords are the abilities this permanent's printed
	// static gives its CONTROLLER — "You have hexproof" (Leyline of
	// Sanctity, Aegis of the Gods). Engine ability tokens, in the
	// vocabulary protection.go and keywords.go already parse. Read
	// from the battlefield through CatalogAbilityKey, never from a
	// card's own zone; see game.CatalogPlayerKeywords and ADR 0072's
	// 2026-09-22 amendment (#1197).
	PlayerKeywords []string
	// PlayerLifeTotalLocked is this permanent's printed "Your life
	// total can't change" (CR 119.7 / CR 119.8 — Platinum Emperion),
	// about its CONTROLLER. Read from the battlefield through
	// CatalogAbilityKey, never from a card's own zone; see
	// game.CatalogPlayerLifeTotalLocked and ADR 0085 (#1200).
	PlayerLifeTotalLocked bool
	// DamageCantBePrevented are this permanent's printed "damage can't
	// be prevented" statics (CR 615.12, ADR 0107 §5). Read from the
	// battlefield through CatalogUnpreventableDamage, keyed by
	// CatalogAbilityKey; see unpreventable_damage.go.
	DamageCantBePrevented []UnpreventableDamageStatic
	// CantGainLife are this permanent's printed "can't gain life"
	// statics (CR 119.7, ADR 0107 §5). Read from the battlefield
	// through CatalogCantGainLife; see cant_gain_life.go.
	CantGainLife []CantGainLifeStatic
	// DamageAsThough are this permanent's printed "damage is dealt as
	// though its source had wither / infect" statics (ADR 0108 §10). Read
	// from the battlefield through CatalogDamageAsThough, keyed by
	// CatalogAbilityKey; see damage_as_though.go.
	DamageAsThough []DamageAsThoughStatic
	// AnyColorSpend are this permanent's printed "you may spend mana as
	// though it were mana of any color" statics (CR 609.4b, #1600) —
	// Chromatic Orrery, Mycosynth Lattice, Oath of Nissa. Read from the
	// battlefield through CatalogAnyColorSpend; see spend_any_color.go.
	AnyColorSpend []AnyColorSpendStatic
	// GameEndGates are this permanent's printed "you can't lose the
	// game" / "your opponents can't win the game" statics (CR 104.3),
	// scoped relative to its CONTROLLER. Read from the battlefield
	// through CatalogAbilityKey; see game.CatalogGameEndGates and
	// ADR 0057 Decision 4 (#749).
	GameEndGates []GameEndGate
	// Emblem is the presentation half of an EMBLEM's catalog entry
	// (CR 114) — its board label and its printed ability text. Set
	// only on an emblem's own def, the one effects.Register files
	// under game.EmblemKey(spec.OracleID), so a non-nil Emblem is how
	// the engine tells an emblem's def from a card's. The emblem's
	// abilities are the ordinary Static and Triggered slots above.
	// See emblem.go and ADR 0064. Added in S40 (#623).
	Emblem *EmblemDef

	// TokenText is the presentation half of a TOKEN TEMPLATE's catalog
	// entry (CR 111.1) — the token's printed ability text, verbatim,
	// as the player reads it on the board. Set only on a token
	// template's own def, the one effects files under
	// game.TokenKey(slug), and only when the token prints something.
	//
	// It is here for the reason EmblemDef.Text is: a token has no
	// printing behind it, so there is no oracle text for the client to
	// fetch and a trigger's Label is a log line, not card text. Read
	// through CatalogTokenText / TokenTextForCard; see token_key.go
	// and ADR 0083.
	TokenText string

	// GrantText is a granted ability BUNDLE's printed text (ADR 0093
	// Decision 8) — the quoted ability as the granting card prints it,
	// "{T}: Add one mana of any color." Set only on a bundle's own def,
	// the one effects files under GrantKey(name), from
	// effects.AbilityGrant.Text.
	//
	// It is here for TokenText's reason: the recipient has no printing
	// that says it has the ability, and a granted trigger has no row on
	// the wire at all, so this string is the only thing that tells a
	// player what their permanent can now do. Read through
	// GrantTextFor.
	GrantText string

	// XMatters says everything the card does scales with the
	// announced X, so X=0 does nothing at all. Read only by the
	// legal-move enumerator, through XMattersFor; see
	// effects.Spec.XMatters and internal/legal/x.go.
	XMatters bool

	// WantsDistinctColors marks a spell that READS the colours of the
	// mana that paid for it — converge (CR 702.86) and sunburst (CR
	// 702.44), and nothing else today. It picks the colour-maximising
	// payment strategy at the cast gate (#761).
	WantsDistinctColors bool

	// WantsManaFrom is the kinds of mana SOURCE this card's text reads
	// back — Treasure, creature, artifact (#1212). Read through
	// CatalogWantsManaFrom; it is an ORDERING HINT for the auto-tapper
	// and never a filter (mana_source.go, autotap.go).
	WantsManaFrom ManaSourceKinds

	// AdditionalLandPlays is how many EXTRA lands per turn this
	// permanent lets its controller play while it is on the
	// battlefield — 1 for Exploration, 2 for Azusa (#500). Read
	// through CatalogAdditionalLandPlays; see land_drops.go.
	AdditionalLandPlays int

	// CastPermissions are the STANDING cast and play permissions this
	// permanent grants its controller while it is on the battlefield
	// (ADR 0066) — Underworld Breach's escape for every nonland card
	// in your graveyard, Bolas's Citadel's top of the library. Read
	// through CatalogCastPermissions; see cast_permission.go.
	//
	// Derived on every query rather than written onto the player, so
	// two sources compose and one leaving cannot revoke the other's
	// permission. Per-INSTANCE permissions (Snapcaster, impulse exile)
	// are not here — they are granted by an effect and stored on the
	// player.
	CastPermissions []CastPermission

	// GatedCastPermissions are standing permissions gated by an ADR
	// 0071 designation and/or a card Condition (#1314) — Fortune
	// Teller's Talent's level-2 line: "as long as this Class is level
	// 2 or greater [ActiveWhen], you may play cards from the top of
	// your library [if you've cast a spell this turn, Condition]".
	//
	// A separate slice from CastPermissions, not a field on
	// CastPermission itself, because CastPermission is ALSO the type
	// stored on Player.CastPermissions and mirrored into GameSnapshot
	// — see CastPermissionGate's own doc comment. Read through
	// CatalogGatedCastPermissions.
	GatedCastPermissions []CastPermissionGate

	// CastTimings are the per-player cast-timing statements this
	// permanent makes while it is on the battlefield (#1195) —
	// Vedalken Orrery's "you may cast spells as though they had
	// flash", Teferi, Time Raveler's "each opponent can cast spells
	// only any time they could cast a sorcery". Read through
	// CatalogCastTimings; see cast_timing.go.
	//
	// Derived on every query rather than written onto a player, for
	// the reason CastPermissions gives one field up. A statement that
	// OUTLIVES its source (Emergence Zone's "this turn") is not here:
	// it is granted by an effect and stored on the player.
	CastTimings []CastTimingRule

	// LibraryTopVisible is how far this permanent makes its
	// controller's top library card visible (CR 401.5) — "you may look
	// at the top card of your library any time" is LibraryTopOwner,
	// "play with the top card of your library revealed" is
	// LibraryTopRevealed. Read through CatalogLibraryTopVisible; see
	// library_top.go.
	LibraryTopVisible LibraryTopVisibility
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
	// CR 707.9a, #665: a key carrying granted-ability names is
	// answered with the card's definition merged with each grant's.
	// This is the ONE place a granted ability becomes findable, which
	// is why every existing reader — the trigger harvest, the layer
	// pass, the activation path, the view — needed no change of its
	// own. See copy_grants.go.
	//
	// Since ADR 0093 the same composite carries LAYER-6 grants too
	// (CatalogAbilityKey), and its base may be empty ("|grant:<a>").
	if strings.IndexByte(key, grantKeySeparator[0]) >= 0 {
		return mergedCatalogDef(key)
	}
	// ADR 0103: a fused split spell's key is synthetic, answered from
	// its two halves' own entries (split_fuse.go).
	if strings.HasSuffix(key, fusedKeySuffix) {
		return fusedCatalogDef(strings.TrimSuffix(key, fusedKeySuffix))
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
	CatalogEntersWithCountersFromCast = func(key string) []EntryCountersFromCast {
		if d := catalogDef(key); d != nil {
			return d.EntersWithCountersFromCast
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
	CatalogManaTriggers = func(key string) []ManaTrigger {
		if d := catalogDef(key); d != nil {
			return d.ManaTriggers
		}
		return nil
	}
	CatalogTriggerDoublers = func(key string) []TriggerDoubler {
		if d := catalogDef(key); d != nil {
			return d.TriggerDoublers
		}
		return nil
	}
	CatalogTriggerSuppressors = func(key string) []TriggerSuppressor {
		if d := catalogDef(key); d != nil {
			return d.TriggerSuppressors
		}
		return nil
	}
	CatalogAdditionalCost = func(key string) *AdditionalCost {
		if d := catalogDef(key); d != nil {
			return d.AdditionalCost
		}
		return nil
	}
	CatalogOptionalCosts = func(key string) []AdditionalCost {
		if d := catalogDef(key); d != nil {
			return d.OptionalCosts
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
	CatalogDelve = func(key string) bool {
		if d := catalogDef(key); d != nil {
			return d.Delve
		}
		return false
	}
	CatalogSpellsHaveDelve = func(key string) bool {
		if d := catalogDef(key); d != nil {
			return d.SpellsYouCastHaveDelve
		}
		return false
	}
	CatalogCostModifiers = func(key string) []CostModifier {
		if d := catalogDef(key); d != nil {
			return d.CostModifiers
		}
		return nil
	}
	CatalogExhaustPermissions = func(key string) []ExhaustPermission {
		if d := catalogDef(key); d != nil {
			return d.ExhaustPermissions
		}
		return nil
	}
	CatalogActivationRestrictions = func(key string) []ActivationRestriction {
		if d := catalogDef(key); d != nil {
			return d.ActivationRestrictions
		}
		return nil
	}
	CatalogActivationTimings = func(key string) []ActivationTiming {
		if d := catalogDef(key); d != nil {
			return d.ActivationTimings
		}
		return nil
	}
	CatalogAttackTaxes = func(key string) []AttackTax {
		if d := catalogDef(key); d != nil {
			return d.AttackTaxes
		}
		return nil
	}
	CatalogBlockRules = func(key string) []BlockRule {
		if d := catalogDef(key); d != nil {
			return d.BlockRules
		}
		return nil
	}
	CatalogAttackLimits = func(key string) []AttackLimit {
		if d := catalogDef(key); d != nil {
			return d.AttackLimits
		}
		return nil
	}
	CatalogHexproofBypasses = func(key string) []HexproofBypass {
		if d := catalogDef(key); d != nil {
			return d.HexproofBypasses
		}
		return nil
	}
	CatalogWardSuppressions = func(key string) []WardSuppression {
		if d := catalogDef(key); d != nil {
			return d.WardSuppressions
		}
		return nil
	}
	CatalogCastableZones = func(key string) []ZoneKind {
		if d := catalogDef(key); d != nil {
			return d.CastableZones
		}
		return nil
	}
	CatalogSpecialActions = func(key string) []SpecialAction {
		if d := catalogDef(key); d != nil {
			return d.SpecialActions
		}
		return nil
	}
	CatalogSpecialActionGrants = func(key string) []SpecialActionGrant {
		if d := catalogDef(key); d != nil {
			return d.SpecialActionGrants
		}
		return nil
	}
	CatalogCastRestrictions = func(key string) []CastRestriction {
		if d := catalogDef(key); d != nil {
			return d.CastRestrictions
		}
		return nil
	}
	CatalogLandPlayRestrictions = func(key string) []LandPlayRestriction {
		if d := catalogDef(key); d != nil {
			return d.LandPlayRestrictions
		}
		return nil
	}
	CatalogUntapStepPermissions = func(key string) []UntapStepPermission {
		if d := catalogDef(key); d != nil {
			return d.UntapStep
		}
		return nil
	}
	CatalogDrawStepPermissions = func(key string) []DrawStepPermission {
		if d := catalogDef(key); d != nil {
			return d.DrawStep
		}
		return nil
	}
	CatalogUntapStepRestrictions = func(key string) []UntapStepRestriction {
		if d := catalogDef(key); d != nil {
			return d.UntapStepRestrictions
		}
		return nil
	}
	CatalogUntapCaps = func(key string) []UntapCap {
		if d := catalogDef(key); d != nil {
			return d.UntapCaps
		}
		return nil
	}
	CatalogUntapOptOuts = func(key string) []UntapOptOut {
		if d := catalogDef(key); d != nil {
			return d.UntapOptOuts
		}
		return nil
	}
	CatalogCantBeCountered = func(key string) bool {
		d := catalogDef(key)
		return d != nil && d.CantBeCountered
	}
	CatalogCantBeCounteredIf = func(key string) SpellCondition {
		if d := catalogDef(key); d != nil {
			return d.CantBeCounteredIf
		}
		return nil
	}
	CatalogSpellDamageUnpreventable = func(key string) SpellCondition {
		if d := catalogDef(key); d != nil {
			return d.SpellDamageCantBePrevented
		}
		return nil
	}
	CatalogCounterShields = func(key string) []CounterShieldStatic {
		if d := catalogDef(key); d != nil {
			return d.SpellsCantBeCountered
		}
		return nil
	}
	CatalogNoMaxHandSize = func(key string) bool {
		d := catalogDef(key)
		return d != nil && d.NoMaxHandSize
	}
	CatalogPlayerKeywords = func(key string) []string {
		if d := catalogDef(key); d != nil {
			return d.PlayerKeywords
		}
		return nil
	}
	CatalogPlayerLifeTotalLocked = func(key string) bool {
		d := catalogDef(key)
		return d != nil && d.PlayerLifeTotalLocked
	}
	CatalogUnpreventableDamage = func(key string) []UnpreventableDamageStatic {
		if d := catalogDef(key); d != nil {
			return d.DamageCantBePrevented
		}
		return nil
	}
	CatalogCantGainLife = func(key string) []CantGainLifeStatic {
		if d := catalogDef(key); d != nil {
			return d.CantGainLife
		}
		return nil
	}
	CatalogDamageAsThough = func(key string) []DamageAsThoughStatic {
		if d := catalogDef(key); d != nil {
			return d.DamageAsThough
		}
		return nil
	}
	CatalogAnyColorSpend = func(key string) []AnyColorSpendStatic {
		if d := catalogDef(key); d != nil {
			return d.AnyColorSpend
		}
		return nil
	}
	CatalogGameEndGates = func(key string) []GameEndGate {
		if d := catalogDef(key); d != nil {
			return d.GameEndGates
		}
		return nil
	}
	CatalogWantsDistinctColors = func(key string) bool {
		d := catalogDef(key)
		return d != nil && d.WantsDistinctColors
	}
	CatalogWantsManaFrom = func(key string) ManaSourceKinds {
		if d := catalogDef(key); d != nil {
			return d.WantsManaFrom
		}
		return 0
	}
	CatalogAdditionalLandPlays = func(key string) int {
		if d := catalogDef(key); d != nil {
			return d.AdditionalLandPlays
		}
		return 0
	}
	CatalogXMatters = func(key string) bool {
		d := catalogDef(key)
		return d != nil && d.XMatters
	}
	CatalogCastPermissions = func(key string) []CastPermission {
		if d := catalogDef(key); d != nil {
			return d.CastPermissions
		}
		return nil
	}
	CatalogGatedCastPermissions = func(key string) []CastPermissionGate {
		if d := catalogDef(key); d != nil {
			return d.GatedCastPermissions
		}
		return nil
	}
	CatalogCastTimings = func(key string) []CastTimingRule {
		if d := catalogDef(key); d != nil {
			return d.CastTimings
		}
		return nil
	}
	CatalogLibraryTopVisible = func(key string) LibraryTopVisibility {
		if d := catalogDef(key); d != nil {
			return d.LibraryTopVisible
		}
		return LibraryTopHidden
	}
	CatalogTokenText = func(key string) string {
		if d := catalogDef(key); d != nil {
			return d.TokenText
		}
		return ""
	}
}
