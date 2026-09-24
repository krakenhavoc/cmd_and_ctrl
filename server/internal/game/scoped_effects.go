package game

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// scoped_effects.go is ADR 0041 phase 3's data record for a continuous
// effect created by a resolving spell or ability (CR 611.2) — see the
// 2026-09-24 amendment to ADR 0041, Decision P1, and #1497.
//
// WHY A SECOND REGISTRY BESIDE ScopedStatics. A `ScopedStatic` holds a
// `StaticAbility`, which is two closures, so a game holding one is not
// a restore point (`ContinuationCensus.ScopedStatics`). An earthbent
// land or an Agent of Treachery's theft keeps one alive for the rest
// of the game, and so froze its table's restore point for the rest of
// the game. A `ScopedEffect` says the same thing as DATA: the objects
// it affects, a list of operations from a closed vocabulary, one
// timestamp and an ADR 0063 `Duration`. The snapshot carries it
// verbatim, and the running binary interprets it.
//
// The measurement behind the vocabulary is in the ADR: every scoped
// static in the tree — 115 card files and every engine site — is one
// of twelve operations over a pinned object set. Nothing needs a
// bespoke closure, so there is no per-card registry here; the
// `ModKind` string is the key, and the engine owns its namespace.
//
// The legacy registry stays until every writer has moved (ADR 0041 P5,
// tier 3a). Both feed the same layer pass, sort together by timestamp
// within each layer bucket (CR 613.7), and are swept by the same
// `durationExpiredLocked`.

// ModKind names one operation a ScopedEffect performs. It is a STRING
// on disk, never an int, so renumbering constants can never change
// what an old restore point means — and it is an on-disk identity like
// a token slug or a grant key: never renamed, never reused. A mod
// whose meaning has to change gets a new kind.
//
// A restore point naming a kind this binary does not know is refused
// with ErrUnknownEffectKey rather than guessed at (ADR 0041 P4), which
// is what lets the vocabulary grow within a schema version.
type ModKind string

// The kinds. The layer each belongs to is fixed by modKinds below, so
// a record cannot put an operation in the wrong layer.
//
// `grantAbilities` (ADR 0093's ScopedGrant) joins this list with
// ADR 0093's PR 4, which is where the layer-6 grant seam it adapts
// into lands; it is not declared before something can interpret it.
const (
	ModSetController    ModKind = "setController"    // layer 2
	ModAddTypes         ModKind = "addTypes"         // layer 4
	ModRemoveTypes      ModKind = "removeTypes"      // layer 4
	ModAddSubtypes      ModKind = "addSubtypes"      // layer 4
	ModAllCreatureTypes ModKind = "allCreatureTypes" // layer 4
	ModSetColors        ModKind = "setColors"        // layer 5
	ModAddKeywords      ModKind = "addKeywords"      // layer 6
	ModRemoveKeywords   ModKind = "removeKeywords"   // layer 6
	ModLoseAllAbilities ModKind = "loseAllAbilities" // layer 6
	ModAddRestrictions  ModKind = "addRestrictions"  // layer 6
	ModSetBasePower     ModKind = "setBasePower"     // layer 7b
	ModSetBaseToughness ModKind = "setBaseToughness" // layer 7b
	ModModifyPT         ModKind = "modifyPT"         // layer 7c
	// ModAddAttackRequirement is a CR 508.1d attack requirement
	// (#1571): "attacks … if able", or — with Player set — "attacks a
	// player other than Player if able". Layer 6 for the reason
	// ModAddRestrictions is: not a characteristic, written where the
	// text sits, only ever appended to (attack_requirements.go).
	ModAddAttackRequirement ModKind = "addAttackRequirement" // layer 6
)

// AffectedScope is a ScopedEffect's affected set as a RULE read live at
// every layer pass, instead of a set of objects locked when the effect
// began (#1571). CR 611.2c locks the set only for an effect that
// changes characteristics or control; an effect that does neither
// "modifies the rules of the game, so it can affect objects that
// weren't affected when that continuous effect began" — Bident of
// Thassa's "creatures your opponents control attack this turn if able"
// reaches a creature an opponent casts after it resolved.
//
// A closed vocabulary like ModKind, and an on-disk identity for the
// same reason: never renamed, never reused. A restore point naming a
// scope this binary does not know is refused (ErrUnknownEffectKey).
type AffectedScope string

const (
	// ScopeNone is the ordinary record: Affected is the set.
	ScopeNone AffectedScope = ""
	// ScopeOpponentsCreatures is "creatures your opponents control",
	// the "you" being the record's Controller.
	ScopeOpponentsCreatures AffectedScope = "opponentsCreatures"
)

// KnownAffectedScope reports whether this binary can interpret s.
func KnownAffectedScope(s AffectedScope) bool {
	return s == ScopeNone || s == ScopeOpponentsCreatures
}

// Mod is one operation of a ScopedEffect, with its plain parameters.
// Which fields a kind reads is documented on its constructor; the rest
// stay zero and are omitted on disk.
//
// PARAMETER TYPES ARE CLOSED (ADR 0041 P1): bool, ints, strings,
// uuid.UUID, the engine's string and bit enums, and slices of those.
// Never a func, an interface, a pointer, a Card or a CardPredicate —
// TestScopedEffectIsPureData holds the line.
type Mod struct {
	Kind         ModKind     `json:"kind"`
	Types        []string    `json:"types,omitempty"`
	Subtypes     []string    `json:"subtypes,omitempty"`
	Colors       []string    `json:"colors,omitempty"`
	Keywords     []string    `json:"keywords,omitempty"`
	Restrictions Restriction `json:"restrictions,omitempty"`
	Power        int         `json:"power,omitempty"`
	Toughness    int         `json:"toughness,omitempty"`
	Player       uuid.UUID   `json:"player,omitempty"`
}

// AffectedObject is one member of a ScopedEffect's affected set: the
// key every CR 611.2c set in the engine uses — an instance plus the
// battlefield-entry stamp it carried when the effect began, so a
// permanent that left and came back is a new object the effect no
// longer follows (CR 400.7).
//
// EnteredAt 0 means "this instance, whatever its entry stamp", the
// convention `sameObjectOnBattlefieldLocked` already uses. It is how
// suspend's haste follows its card from the stack onto the battlefield
// (CR 702.62a), and it is safe only because that record's Duration
// ends the effect the moment the card stops being the object it named.
// Nothing else may write it: pin a permanent with PinObject.
//
// Unstamped is the other side of that line (#1558): a permanent that
// carries no entry stamp — a fixture's seeded card, one put onto the
// battlefield without the zone-move event — pinned EXACTLY. It matches
// that instance only while it is still unstamped, so a flicker, which
// stamps the new object, ends it (CR 400.7), the way the legacy
// closure's `== stamp` comparison always did. Before #1558 such a
// permanent was written as the wildcard and stayed matched. A new
// field rather than a sentinel stamp, so an older v7 binary refuses a
// file carrying one (ADR 0041 P4) instead of reading it as a stamp no
// object has.
//
// Controller, when set, narrows the member further: the object is
// affected only while that player controls it. Suspend's haste again —
// "it has haste until that player loses control of it".
type AffectedObject struct {
	ID         uuid.UUID `json:"id"`
	EnteredAt  int64     `json:"enteredAt,omitempty"`
	Unstamped  bool      `json:"unstamped,omitempty"`
	Controller uuid.UUID `json:"controller,omitempty"`
}

// PinObject is the affected-set member for the permanent `id` that
// entered the battlefield at `stamp`: pinned to that stamp, or — for an
// unstamped permanent — to being unstamped. Never the wildcard.
func PinObject(id uuid.UUID, stamp int64) AffectedObject {
	return AffectedObject{ID: id, EnteredAt: stamp, Unstamped: stamp == 0}
}

// entryMatches reports whether a permanent whose entry stamp is now
// `stamp` is the object an (enteredAt, unstamped) pin names: an exact
// "still unstamped" when unstamped is set, otherwise the 0 wildcard or
// an exact stamp. The one reading of a pin, shared by the affected set
// and Duration.Pinned.
func entryMatches(enteredAt int64, unstamped bool, stamp int64) bool {
	if unstamped {
		return stamp == 0
	}
	return enteredAt == 0 || stamp == enteredAt
}

// ScopedEffect is a continuous effect created by a resolving spell or
// ability (CR 611.2), as data. See the file comment.
//
// IMMUTABILITY CONTRACT, the same one ScopedStatic keeps: every field
// is written once at registration and never mutated. Clone copies the
// slice into a fresh backing array and shares the inner slices.
type ScopedEffect struct {
	// Affected is the CR 611.2c set, locked when the effect began.
	Affected []AffectedObject `json:"affected"`

	// Scope, when set, replaces Affected with a rule read live at
	// every pass (#1571): the record affects whatever the rule matches
	// now, including objects that did not exist when it began. Only
	// for an effect that changes neither characteristics nor control
	// (CR 611.2c) — today, an attack requirement.
	Scope AffectedScope `json:"scope,omitempty"`

	// Mods are applied each in its own layer, all at Timestamp. One
	// record may span several layers — earthbend is layer 4, 6 and 7b
	// — and that is what makes it ONE effect for CR 613.7.
	Mods []Mod `json:"mods"`

	// Source is the spell or ability's source object, as an identity:
	// what `ControlSource` names for a control change (#930), and what
	// an operator reads. Last-known information is enough.
	Source     ObjectRef `json:"source"`
	SourceName string    `json:"sourceName,omitempty"`

	// Controller is who controlled the source as the effect began —
	// the "you" of the effect, read once (CR 611.2c).
	Controller uuid.UUID `json:"controller,omitempty"`

	// Timestamp is the CR 613.7 timestamp. The two halves of an
	// exchange of control share one (CR 701.12).
	Timestamp int64 `json:"timestamp"`

	// Duration is how long the effect lasts (ADR 0063), swept by
	// `durationExpiredLocked` like every other duration in the game.
	Duration Duration `json:"duration"`

	// Label is human-readable attribution for logs and tests.
	Label string `json:"label,omitempty"`
}

// modKindSpec is where a kind lives in the layer system.
type modKindSpec struct {
	layer    Layer
	subLayer SubLayer
	removes  bool // CR 613.1f ability removal (ADR 0046)
}

// modKinds is the closed vocabulary. A kind missing here is unknown:
// registration refuses it and restore refuses a file naming it.
var modKinds = map[ModKind]modKindSpec{
	ModSetController:    {layer: Layer2Control},
	ModAddTypes:         {layer: Layer4Type},
	ModRemoveTypes:      {layer: Layer4Type},
	ModAddSubtypes:      {layer: Layer4Type},
	ModAllCreatureTypes: {layer: Layer4Type},
	ModSetColors:        {layer: Layer5Color},
	ModAddKeywords:      {layer: Layer6Ability},
	ModRemoveKeywords:   {layer: Layer6Ability},
	ModLoseAllAbilities: {layer: Layer6Ability, removes: true},
	ModAddRestrictions:  {layer: Layer6Ability},
	ModSetBasePower:     {layer: Layer7PT, subLayer: SubLayer7B_Set},
	ModSetBaseToughness: {layer: Layer7PT, subLayer: SubLayer7B_Set},
	ModModifyPT:         {layer: Layer7PT, subLayer: SubLayer7C_Modify},
	// #1571
	ModAddAttackRequirement: {layer: Layer6Ability},
}

// KnownModKind reports whether this binary can interpret k.
func KnownModKind(k ModKind) bool {
	_, ok := modKinds[k]
	return ok
}

// ModKinds is the vocabulary, sorted — for tests and diagnostics.
func ModKinds() []ModKind {
	out := make([]ModKind, 0, len(modKinds))
	for k := range modKinds {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// ---------------------------------------------------------------
// Constructors — what effect code writes
// ---------------------------------------------------------------

// SetControllerMod is "that permanent is controlled by <player>"
// (layer 2, CR 613.1b). Reads Player.
func SetControllerMod(player uuid.UUID) Mod { return Mod{Kind: ModSetController, Player: player} }

// AddTypesMod adds card types in addition to the object's others
// (layer 4). Reads Types.
func AddTypesMod(types ...string) Mod { return Mod{Kind: ModAddTypes, Types: copyStrings(types)} }

// RemoveTypesMod removes card types (layer 4) — Enduring Curiosity's
// "it's not a creature". Reads Types.
func RemoveTypesMod(types ...string) Mod {
	return Mod{Kind: ModRemoveTypes, Types: copyStrings(types)}
}

// AddSubtypesMod adds subtypes in addition to the object's others
// (layer 4) — amass's "it's also an Orc", Kyoshi's Island. Reads
// Subtypes.
func AddSubtypesMod(subtypes ...string) Mod {
	return Mod{Kind: ModAddSubtypes, Subtypes: copyStrings(subtypes)}
}

// AllCreatureTypesMod is "is every creature type" (layer 4,
// Characteristic.AllCreatureTypes).
func AllCreatureTypesMod() Mod { return Mod{Kind: ModAllCreatureTypes} }

// SetColorsMod sets the object's colours (layer 5). Reads Colors;
// an empty list makes it colourless.
func SetColorsMod(colors ...string) Mod { return Mod{Kind: ModSetColors, Colors: copyStrings(colors)} }

// AddKeywordsMod grants keyword abilities (layer 6) through
// AppendKeywordAbility, so a cumulative keyword stays cumulative and
// any other is not duplicated. Reads Keywords.
func AddKeywordsMod(keywords ...string) Mod {
	return Mod{Kind: ModAddKeywords, Keywords: copyStrings(keywords)}
}

// RemoveKeywordsMod removes keyword abilities (layer 6). Reads
// Keywords.
func RemoveKeywordsMod(keywords ...string) Mod {
	return Mod{Kind: ModRemoveKeywords, Keywords: copyStrings(keywords)}
}

// LoseAllAbilitiesMod is "loses all abilities" (layer 6, a CR 613.1f
// removal — ADR 0046), appending `keep` back as the same effect's
// grant. Reads Keywords.
func LoseAllAbilitiesMod(keep ...string) Mod {
	return Mod{Kind: ModLoseAllAbilities, Keywords: copyStrings(keep)}
}

// AddRestrictionsMod adds restriction bits (layer 6, ADR 0045). Reads
// Restrictions.
func AddRestrictionsMod(r Restriction) Mod { return Mod{Kind: ModAddRestrictions, Restrictions: r} }

// SetBasePowerMod sets base power (layer 7b). Reads Power.
func SetBasePowerMod(n int) Mod { return Mod{Kind: ModSetBasePower, Power: n} }

// SetBaseToughnessMod sets base toughness (layer 7b). Reads Toughness.
func SetBaseToughnessMod(n int) Mod { return Mod{Kind: ModSetBaseToughness, Toughness: n} }

// AddAttackRequirementMod is a CR 508.1d attack requirement (#1571):
// "attacks … if able" with otherThan uuid.Nil, "attacks a player other
// than otherThan if able" with it set. The requirement is attributed to
// the record's source, so a refusal names the card.
func AddAttackRequirementMod(otherThan uuid.UUID) Mod {
	return Mod{Kind: ModAddAttackRequirement, Player: otherThan}
}

// SetBasePTMods is "has base power and toughness P/T" — both 7b
// halves.
func SetBasePTMods(power, toughness int) []Mod {
	return []Mod{SetBasePowerMod(power), SetBaseToughnessMod(toughness)}
}

// ModifyPTMod is "gets +P/+T" (layer 7c). Reads Power and Toughness;
// negative values shrink.
func ModifyPTMod(power, toughness int) Mod {
	return Mod{Kind: ModModifyPT, Power: power, Toughness: toughness}
}

// ---------------------------------------------------------------
// Registration
// ---------------------------------------------------------------

// PinnedObjectsLocked returns the affected set for the given
// battlefield permanents, each keyed on the entry stamp it carries
// now (PinObject: an unstamped permanent is pinned exactly, not as the
// wildcard — #1558). IDs not on the battlefield are skipped, so the
// result may be empty. Caller must hold g.mu.
func (g *Game) PinnedObjectsLocked(ids ...uuid.UUID) []AffectedObject {
	out := make([]AffectedObject, 0, len(ids))
	for _, id := range ids {
		if c, ok := g.battlefieldCardLocked(id); ok {
			out = append(out, PinObject(id, c.EnteredBattlefieldAt))
		}
	}
	return out
}

// RegisterScopedEffectForEffect installs a data-backed continuous
// effect and bumps the layer version so the next recompute applies it.
// It reports false, having registered nothing, when there is nothing
// to affect or nothing to do.
//
// `sourceID` names the spell or ability's source; it is looked up for
// its identity, name and controller, and a miss stores the bare ID.
//
// It PANICS on a mod kind this binary does not know: that is a
// programming error in the caller, caught by the first test that
// exercises it, and never a state a player can reach.
//
// Caller must hold g.mu (write). Effects call this from inside the
// resolution frame, which already holds it.
func (g *Game) RegisterScopedEffectForEffect(sourceID uuid.UUID, affected []AffectedObject, mods []Mod, d Duration, label string) bool {
	return g.registerScopedEffectLocked(sourceID, affected, mods, d, label, timeNowUnixNano())
}

// registerScopedEffectLocked is the shared body with the CR 613.7
// timestamp passed in — chosen rather than read only for the two
// halves of an exchange (CR 701.12). Caller must hold g.mu (write).
func (g *Game) registerScopedEffectLocked(sourceID uuid.UUID, affected []AffectedObject, mods []Mod, d Duration, label string, ts int64) bool {
	if len(affected) == 0 || len(mods) == 0 {
		return false
	}
	return g.appendScopedEffectLocked(sourceID, affected, ScopeNone, uuid.Nil, mods, d, label, ts)
}

// RegisterScopedRuleEffectForEffect installs a record whose affected
// set is a live RULE (AffectedScope) rather than a locked set of
// objects — CR 611.2c's "modifies the rules of the game" case, #1571.
// `controller` is the effect's "you" (the controller of the resolving
// spell or ability), which the scope reads. Reports false, registering
// nothing, for an unknown scope or no mods.
//
// Caller must hold g.mu (write).
func (g *Game) RegisterScopedRuleEffectForEffect(sourceID uuid.UUID, scope AffectedScope, controller uuid.UUID, mods []Mod, d Duration, label string) bool {
	if scope == ScopeNone || !KnownAffectedScope(scope) || len(mods) == 0 {
		return false
	}
	return g.appendScopedEffectLocked(sourceID, nil, scope, controller, mods, d, label, timeNowUnixNano())
}

// appendScopedEffectLocked is the one body both registrations share.
// A non-nil `controller` overrides the one read off the source.
//
// Caller must hold g.mu (write).
func (g *Game) appendScopedEffectLocked(sourceID uuid.UUID, affected []AffectedObject, scope AffectedScope, controller uuid.UUID, mods []Mod, d Duration, label string, ts int64) bool {
	for _, m := range mods {
		if !KnownModKind(m.Kind) {
			panic(fmt.Sprintf("game: scoped effect %q uses unknown mod kind %q", label, m.Kind))
		}
	}
	e := ScopedEffect{
		Affected:  append([]AffectedObject(nil), affected...),
		Scope:     scope,
		Mods:      cloneMods(mods),
		Source:    ObjectRef{ID: sourceID},
		Timestamp: ts,
		Duration:  d,
		Label:     label,
	}
	if src, ok := g.LookupCardForEffect(sourceID); ok {
		e.Source.Epoch = src.ObjectEpoch
		e.SourceName = src.Name
		e.Controller = src.Controller
	}
	if controller != uuid.Nil {
		e.Controller = controller
	}
	g.ScopedEffects = append(g.ScopedEffects, e)
	g.layerVersion.Add(1)
	return true
}

func cloneMods(mods []Mod) []Mod {
	if len(mods) == 0 {
		return nil
	}
	out := make([]Mod, len(mods))
	for i, m := range mods {
		m.Types = copyStrings(m.Types)
		m.Subtypes = copyStrings(m.Subtypes)
		m.Colors = copyStrings(m.Colors)
		m.Keywords = copyStrings(m.Keywords)
		out[i] = m
	}
	return out
}

// cloneScopedEffects copies the registry into a fresh backing array.
// Entries are immutable, so their inner slices are shared, exactly as
// the ScopedStatics clone does.
func cloneScopedEffects(in []ScopedEffect) []ScopedEffect {
	if len(in) == 0 {
		return nil
	}
	out := make([]ScopedEffect, len(in))
	copy(out, in)
	return out
}

// ---------------------------------------------------------------
// Sweep
// ---------------------------------------------------------------

// sweepScopedEffectsLocked drops records whose duration has run out,
// through the one function that decides what a duration means. Same
// fresh-slice rule as sweepScopedStaticsLocked, for the same reason:
// the backing array is shared with every undo snapshot.
//
// Returns whether it dropped anything; the caller bumps the layer
// version. Caller must hold g.mu.
func (g *Game) sweepScopedEffectsLocked(endOfTurn bool) bool {
	if len(g.ScopedEffects) == 0 {
		return false
	}
	kept := make([]ScopedEffect, 0, len(g.ScopedEffects))
	for _, e := range g.ScopedEffects {
		if !g.durationExpiredLocked(e.Duration, endOfTurn) {
			kept = append(kept, e)
		}
	}
	if len(kept) == len(g.ScopedEffects) {
		return false
	}
	if len(kept) == 0 {
		kept = nil
	}
	g.ScopedEffects = kept
	return true
}

// ---------------------------------------------------------------
// The layer-pass adapter
// ---------------------------------------------------------------

// scopedEffectContinuousEffectsLocked adapts every live record into
// one ContinuousEffect per mod. The closures are built here, by the
// running binary, from the record's data — so they are REBUILT on
// every pass and never persisted, which is the whole point.
//
// MEMOISED (#1558). Rebuilding a stub source, a predicate and one
// closure per mod on every recompute cost 100 records on one creature
// about 300 allocations per pass over the legacy closure statics they
// replace — before tier 3a moves Giant Growth and crew onto this path.
// The adaptation depends on nothing but the records (every closure
// reads the board it is handed at call time), so the output is reused
// for as long as the registry holds the same records in the same
// order: see scopedEffectAdapterMemo for what "the same" means and why
// it is record identity rather than the layer version.
//
// The slice returned is capped at its length, so a caller that appends
// to it — activeStaticAbilitiesLocked does — copies rather than
// writing into the memo's backing array.
//
// Caller must hold g.mu in write mode (the recompute pass does).
func (g *Game) scopedEffectContinuousEffectsLocked() []ContinuousEffect {
	if len(g.ScopedEffects) == 0 {
		g.scopedEffectMemo = scopedEffectAdapterMemo{}
		return nil
	}
	if !g.scopedEffectMemo.matches(g.ScopedEffects) {
		g.scopedEffectMemo = scopedEffectAdapterMemo{
			keys:    scopedEffectKeys(g.ScopedEffects),
			effects: adaptScopedEffects(g.ScopedEffects),
		}
	}
	out := g.scopedEffectMemo.effects
	return out[:len(out):len(out)]
}

// scopedEffectAdapterMemo is the adapter's output and the identity of
// every record it was built from, position by position.
//
// WHY IDENTITY AND NOT layerVersion. The layer version is bumped by
// every registration and sweep, but it is not monotone: an undo stores
// the snapshot's version plus one (restoreFromLocked), so the same
// number can name two different registries, and a memo keyed on it
// could hand back a record the undo took away. A record's identity
// cannot collide that way. Records are immutable once registered and
// every registration and restore allocates its Mods afresh
// (cloneMods, deepCopyScopedEffects), so &Mods[0] names one record for
// as long as anything points at it — and the memo's own key keeps it
// alive, so the address cannot be reused while the memo holds it.
// Clone shares the inner slices, which is fine: the same address is
// the same immutable record, in either game.
//
// A cloned Game does not copy the memo (cloneLocked lists its fields
// and this is not one of them), and nothing mutates a memo in place —
// a rebuild assigns new slices — so two games can never write through
// a shared one.
type scopedEffectAdapterMemo struct {
	keys    []scopedEffectMemoKey
	effects []ContinuousEffect
}

// scopedEffectMemoKey is one record's identity. The pointers do the
// work; the scalars are the fields the adapter reads that are not
// behind them, compared so a key is never trusted on the pointers
// alone.
type scopedEffectMemoKey struct {
	mods       *Mod
	affected   *AffectedObject
	nMods      int
	nAffected  int
	scope      AffectedScope
	controller uuid.UUID
	source     uuid.UUID
	sourceName string
	timestamp  int64
}

func scopedEffectKeyOf(e *ScopedEffect) scopedEffectMemoKey {
	k := scopedEffectMemoKey{
		nMods:      len(e.Mods),
		nAffected:  len(e.Affected),
		scope:      e.Scope,
		controller: e.Controller,
		source:     e.Source.ID,
		sourceName: e.SourceName,
		timestamp:  e.Timestamp,
	}
	if len(e.Mods) > 0 {
		k.mods = &e.Mods[0]
	}
	if len(e.Affected) > 0 {
		k.affected = &e.Affected[0]
	}
	return k
}

func scopedEffectKeys(records []ScopedEffect) []scopedEffectMemoKey {
	out := make([]scopedEffectMemoKey, len(records))
	for i := range records {
		out[i] = scopedEffectKeyOf(&records[i])
	}
	return out
}

// matches reports whether the memo was built from exactly these
// records, in this order.
func (m *scopedEffectAdapterMemo) matches(records []ScopedEffect) bool {
	if m.keys == nil || len(m.keys) != len(records) {
		return false
	}
	for i := range records {
		if m.keys[i] != scopedEffectKeyOf(&records[i]) {
			return false
		}
	}
	return true
}

// adaptScopedEffects is the adaptation itself: one ContinuousEffect
// per mod, every closure built from the record's data.
func adaptScopedEffects(records []ScopedEffect) []ContinuousEffect {
	n := 0
	for i := range records {
		n += len(records[i].Mods)
	}
	out := make([]ContinuousEffect, 0, n)
	for i := range records {
		e := records[i]
		// A stub source: the one reader is setController's
		// ControlSource, which wants the instance ID.
		src := &Card{InstanceID: e.Source.ID, Name: e.SourceName, Controller: e.Controller}
		applies := affectedPredicate(e.Affected)
		if e.Scope != ScopeNone {
			applies = scopePredicate(e.Scope, e.Controller)
		}
		for _, m := range e.Mods {
			spec, ok := modKinds[m.Kind]
			if !ok {
				continue // unreachable: registration and restore both refuse it
			}
			out = append(out, staticContinuousEffect{
				ability: StaticAbility{
					Layer:            spec.layer,
					SubLayer:         spec.subLayer,
					RemovesAbilities: spec.removes,
					AppliesTo:        applies,
					Apply:            modApply(m),
				},
				source:    src,
				timestamp: e.Timestamp,
			})
		}
	}
	return out
}

// affectedPredicate is the AppliesTo for a pinned set.
func affectedPredicate(set []AffectedObject) func(*Card, *Game, *Card) bool {
	return func(target *Card, _ *Game, _ *Card) bool {
		if target == nil {
			return false
		}
		for _, a := range set {
			if a.ID != target.InstanceID {
				continue
			}
			if !entryMatches(a.EnteredAt, a.Unstamped, target.EnteredBattlefieldAt) {
				continue
			}
			if a.Controller != uuid.Nil && target.Controller != a.Controller {
				continue
			}
			return true
		}
		return false
	}
}

// scopePredicate is the AppliesTo for a live-rule record (#1571). An
// unknown scope matches nothing — unreachable, since registration and
// restore both refuse one.
func scopePredicate(scope AffectedScope, controller uuid.UUID) func(*Card, *Game, *Card) bool {
	switch scope {
	case ScopeOpponentsCreatures:
		return func(target *Card, _ *Game, _ *Card) bool {
			return target != nil && target.IsCreature() && target.Controller != controller
		}
	}
	return func(*Card, *Game, *Card) bool { return false }
}

// modApply is the interpreter: what each kind does to a
// Characteristic. The only code in the engine that gives a ModKind its
// meaning.
func modApply(m Mod) func(*Characteristic, *Card, *Game, *Card) {
	switch m.Kind {
	case ModSetController:
		player := m.Player
		return func(ch *Characteristic, _ *Card, _ *Game, src *Card) {
			ch.Controller = player
			// #930: whoever wrote Controller last in this bucket won
			// CR 613.7, and the control-change event names it.
			if src != nil {
				ch.ControlSource = src.InstanceID
			}
		}
	case ModAddTypes:
		types := m.Types
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
			for _, t := range types {
				if !typeListHas(ch.Types, t) {
					ch.Types = append(ch.Types, t)
				}
			}
		}
	case ModRemoveTypes:
		types := m.Types
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
			kept := make([]string, 0, len(ch.Types))
			for _, t := range ch.Types {
				if !typeListHas(types, t) {
					kept = append(kept, t)
				}
			}
			ch.Types = kept
		}
	case ModAddSubtypes:
		subtypes := m.Subtypes
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
			for _, s := range subtypes {
				if !typeListHas(ch.Subtypes, s) {
					ch.Subtypes = append(ch.Subtypes, s)
				}
			}
		}
	case ModAllCreatureTypes:
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
			ch.AllCreatureTypes = true
		}
	case ModSetColors:
		colors := m.Colors
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
			ch.Colors = copyStrings(colors)
		}
	case ModAddKeywords, ModLoseAllAbilities:
		// loseAllAbilities: the engine empties the list for a
		// RemovesAbilities effect before Apply runs (ADR 0046), so
		// Apply only appends what the same effect grants back.
		keywords := m.Keywords
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
			for _, kw := range keywords {
				ch.Abilities = AppendKeywordAbility(ch.Abilities, kw)
			}
		}
	case ModRemoveKeywords:
		keywords := m.Keywords
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
			kept := make([]string, 0, len(ch.Abilities))
			for _, a := range ch.Abilities {
				drop := false
				for _, kw := range keywords {
					if strings.EqualFold(a, kw) {
						drop = true
						break
					}
				}
				if !drop {
					kept = append(kept, a)
				}
			}
			ch.Abilities = kept
		}
	case ModAddRestrictions:
		bits := m.Restrictions
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
			ch.Restrictions |= bits
		}
	case ModSetBasePower:
		n := m.Power
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) { ch.Power = n }
	case ModSetBaseToughness:
		n := m.Toughness
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) { ch.Toughness = n }
	case ModAddAttackRequirement:
		otherThan := m.Player
		return func(ch *Characteristic, _ *Card, _ *Game, src *Card) {
			r := AttackRequirement{OtherThan: otherThan}
			if src != nil {
				r.Source, r.SourceName = src.InstanceID, src.Name
			}
			ch.AttackRequirements = append(ch.AttackRequirements, r)
		}
	case ModModifyPT:
		p, t := m.Power, m.Toughness
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
			ch.Power += p
			ch.Toughness += t
		}
	}
	return nil
}
