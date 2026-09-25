package game

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// ability_ref.go is ADR 0041 phase 3's tier 4, first slice (Decision
// P9, #1497): an ability's stack item names the catalog ROW it came
// from, as data, so a table with an ability waiting on the stack is
// still a restore point.
//
// # What is stamped
//
// ActivateCatalogAbility copies the row's Effect, target clause and
// mode clause onto the item verbatim, and nothing is captured at
// activation time — so the row IS the ability, and naming it is enough.
// The item gets two stamps beside what it already holds:
//
//	Body:           "catalog/activated"
//	Params.Ability: {key, slot, ref, name}
//
// The body is the REFUSAL token (P10): no v7 binary before this slice
// registers it, so every one of them refuses a file that names it
// (ErrUnknownEffectKey) and keeps the file for the roll-forward. The
// reference is what restore reads.
//
// Triggered abilities are stamped the same way since slice 4-2, but
// only a row that DECLARES its Effect (TriggeredAbility.Effect): the
// engine builds that item itself, so the row is the ability exactly as
// an activated one is. It gets
//
//	Body:           "catalog/triggered"
//	Params.Ability: {key, "triggered", ref, name}
//
// where the key and ref are the row's catalog identity, stamped by the
// registry when it filed the definition (IdentifyCatalogRows), and the
// name is the row's Key — the stack label every constructor declares. A
// row with a hand-written Build and no Effect may have captured
// trigger-time values in its closure, so it is NOT stamped: its item
// stays an unkeyed closure, the census counts it, and the row is on
// cards/effects/testdata/legacy_trigger_builds.txt until the tail of
// tier 4 converts it. The same goes for a row whose TargetsFrom reads
// the board (TriggeredAbility.TargetsFromReadsBoard): restore could not
// re-derive its clause exactly.
//
// A binary from 4-1 does not register "catalog/triggered", so it
// refuses a file from this slice rather than restoring a trigger it
// cannot rebuild (P10).
//
// # Restore, and the owner's answers to Q2 and Q3 (2026-09-25)
//
// Restore looks the row up once and re-derives the item's Effect,
// target clause and mode clause (every ModeDoing bullet with it) from
// it — the running binary's meaning wins across a deploy, as it does
// for a spell restored by oracle ID.
//
//   - Q2: the ref is checked against the declared label. A row at the
//     ref with the same label is the ability. Otherwise the single row
//     in the same list with that label is, and the ref is rewritten to
//     it (a deploy reordered the card's abilities).
//   - Q3: no row, or more than one, with that label. The game is
//     restored anyway. The item stays on the stack with no automatic
//     effect — a manual item, exactly as a sandbox-announced one is —
//     the source card is flagged AbilitiesLostOnRestore, and the boot
//     path logs one ERROR line (GameSnapshot.LostStackAbilities). The
//     item is never dropped and the game is never abandoned for it.

// AbilityRef names one row of one catalog definition (ADR 0041 P9). It
// rides in EffectParams, so P4's refusal of unknown fields covers it.
type AbilityRef struct {
	// Key is the catalog key whose row list Ref indexes: the object's
	// own key (CatalogKey, copy grants included) for an own:<i> ref,
	// or the bundle's GrantKey for a grant:<bundle>:<i>:<n> ref.
	Key string `json:"key"`
	// Slot is which list of the definition: "activated" or
	// "triggered". This binary refuses any slot it does not read.
	Slot string `json:"slot"`
	// Ref is ADR 0093 Decision 5's row ref, counted in the full
	// declared list before designation gating.
	Ref string `json:"ref"`
	// Name is the row's declared label at announce — an activated
	// row's Label, a triggered row's Key — checked on restore (the
	// owner's answer to Q2).
	Name string `json:"name"`
}

// AbilitySlotActivated is the activated-ability list of a CardDef.
const AbilitySlotActivated = "activated"

// AbilitySlotTriggered is the triggered-ability list of a CardDef
// (tier 4-2).
const AbilitySlotTriggered = "triggered"

// knownAbilitySlot reports whether this binary can resolve a ref in
// the slot.
func knownAbilitySlot(slot string) bool {
	return slot == AbilitySlotActivated || slot == AbilitySlotTriggered
}

// CatalogActivatedBodyKey is the body an activated ability's stack item
// names. It is the refusal token of P10: see the file comment.
const CatalogActivatedBodyKey = "catalog/activated"

// CatalogTriggeredBodyKey is the body a declared triggered ability's
// stack item names (tier 4-2). Like its sibling it is a refusal token:
// no binary before 4-2 registers it.
const CatalogTriggeredBodyKey = "catalog/triggered"

// catalogBodySlot reports whether body is one of the catalog/* bodies,
// and which slot of a CardDef its ref must name.
func catalogBodySlot(body string) (string, bool) {
	switch body {
	case CatalogActivatedBodyKey:
		return AbilitySlotActivated, true
	case CatalogTriggeredBodyKey:
		return AbilitySlotTriggered, true
	}
	return "", false
}

func init() {
	// Resolution through the body is the fallback: the announce path
	// puts the row's Effect on the item directly, and restore
	// re-derives it once, with the name check. A caller that only has
	// the key and the params still resolves through the running
	// binary's row.
	DelayedBody(CatalogActivatedBodyKey, func(g *Game, item *StackItem, p EffectParams) error {
		if p.Ability == nil {
			return fmt.Errorf("%w: %s with no ability ref", ErrUnknownEffectKey, CatalogActivatedBodyKey)
		}
		row, _, outcome := resolveActivatedAbilityRef(*p.Ability)
		if outcome == abilityRefLost {
			return fmt.Errorf("%w: activated ability %s %q of %q", ErrUnknownEffectKey,
				p.Ability.Ref, p.Ability.Name, p.Ability.Key)
		}
		if row.Effect == nil {
			return nil
		}
		return row.Effect(g, item)
	})
	DelayedBody(CatalogTriggeredBodyKey, func(g *Game, item *StackItem, p EffectParams) error {
		if p.Ability == nil {
			return fmt.Errorf("%w: %s with no ability ref", ErrUnknownEffectKey, CatalogTriggeredBodyKey)
		}
		row, _, outcome := resolveTriggeredAbilityRef(*p.Ability)
		if outcome == abilityRefLost {
			return fmt.Errorf("%w: triggered ability %s %q of %q", ErrUnknownEffectKey,
				p.Ability.Ref, p.Ability.Name, p.Ability.Key)
		}
		if row.Effect == nil {
			return nil
		}
		return row.Effect(g, item)
	})
}

// catalogRowID is where the registry filed one triggered row: the key
// whose Triggered list holds it, its index there (before designation
// gating, so a Class levelling up does not renumber it), and — for a
// row merged in from a granted bundle that the object carries more
// than once — which instance of the bundle it came from (ADR 0093 D5's
// <n>). The zero value is "not a catalog row".
type catalogRowID struct {
	key        string
	index      int
	occurrence int
}

// IdentifyCatalogRows stamps every triggered row of a definition with
// its catalog identity: `key` and its index in d.Triggered. The catalog
// calls it once per definition it files — a card, a face, an emblem, a
// token template, a granted bundle — before the definition is readable,
// so every copy of a row the engine hands out afterwards (through
// CatalogTriggers, a merged composite key, a designation filter) knows
// where it came from without anyone threading an index through the
// harvest (ADR 0041 P9, tier 4-2).
//
// The slice is copied, so a Spec the catalog keeps is not written to.
func IdentifyCatalogRows(key string, d *CardDef) {
	if d == nil || key == "" || len(d.Triggered) == 0 {
		return
	}
	rows := make([]TriggeredAbility, len(d.Triggered))
	copy(rows, d.Triggered)
	for i := range rows {
		rows[i].row = catalogRowID{key: key, index: i}
	}
	d.Triggered = rows
}

// numberTriggerRowOccurrences gives each repeat of the same catalog row
// in a merged list its instance number: a composite key naming one
// granted bundle twice (two Cryptolith-style grants of the same bundle)
// has the bundle's rows twice, and the second copy is instance 1. The
// input is returned unchanged — no copy — when nothing repeats, which
// is every list but that one.
func numberTriggerRowOccurrences(rows []TriggeredAbility) []TriggeredAbility {
	var seen map[catalogRowID]int
	var out []TriggeredAbility
	for i := range rows {
		id := rows[i].row
		if id.key == "" {
			continue
		}
		id.occurrence = 0
		if seen == nil {
			seen = make(map[catalogRowID]int, len(rows))
		}
		n := seen[id]
		seen[id] = n + 1
		if n == 0 {
			continue
		}
		if out == nil {
			out = append([]TriggeredAbility(nil), rows...)
		}
		out[i].row.occurrence = n
	}
	if out == nil {
		return rows
	}
	return out
}

// triggeredAbilityRefFor names the catalog row a harvested trigger came
// from, or nil when its item cannot be rebuilt from a row:
//
//   - the row was not filed by the catalog (an engine trigger, a test
//     stub, a reflexive trigger);
//   - it declares no Effect — a hand-written Build that may have
//     captured trigger-time values, the legacy shape;
//   - its TargetsFrom reads the board, so the clause could not be
//     re-derived exactly on the restored board;
//   - the running catalog does not hand the same row back under the
//     ref (a stub, or a row that was never registered under its key).
//
// A nil ref leaves the item an unkeyed closure, counted by the census.
func triggeredAbilityRefFor(t TriggeredAbility) *AbilityRef {
	return TriggeredAbilityRef(t)
}

// TriggeredAbilityRef is the ref a harvested item of this row would be
// stamped with, or nil when the row cannot be named (see
// triggeredAbilityRefFor). Exported for the catalog's own lint, which
// holds every declared row to being nameable.
func TriggeredAbilityRef(t TriggeredAbility) *AbilityRef {
	if t.row.key == "" || t.Effect == nil || t.TargetsFromReadsBoard {
		return nil
	}
	ref := AbilityRef{Key: t.row.key, Slot: AbilitySlotTriggered, Name: t.Key}
	if strings.HasPrefix(t.row.key, GrantKeyPrefix) {
		ref.Ref = GrantedAbilityRef(t.row.key, t.row.index, t.row.occurrence)
	} else {
		ref.Ref = OwnAbilityRef(t.row.index)
	}
	if _, _, outcome := resolveTriggeredAbilityRef(ref); outcome != abilityRefMatched {
		return nil
	}
	return &ref
}

// activatedAbilityRefFor names the catalog row an activation used, or
// nil when the row has no catalog identity — an instance-carried
// ability (census:IntrinsicAbilityCards), or a catalog stub that does
// not hand the same row back. A nil ref leaves the item as it always
// was: an unkeyed closure, counted by the census.
func activatedAbilityRefFor(source Card, index int, origins AbilityOrigins, ab ActivatedAbilityShape) *AbilityRef {
	o := origins.At(index)
	ref := AbilityRef{Slot: AbilitySlotActivated, Ref: o.Ref, Name: ab.Label}
	switch {
	case o.Granted():
		ref.Key = o.Grant
	case strings.HasPrefix(o.Ref, abilityRefOwn):
		// An own row is a catalog row only when the object reads its
		// abilities from the catalog: a card carrying its own
		// instance list (census:IntrinsicAbilityCards) indexes that.
		if source.HasLostAllAbilities() || len(source.ActivatedAbilities) > 0 {
			return nil
		}
		ref.Key = CatalogKey(source)
	default:
		return nil
	}
	if ref.Key == "" {
		return nil
	}
	// The stamp is only as good as the lookup: if the running catalog
	// does not hand this row back under the ref, restore could not
	// either, so do not claim it can.
	if _, _, outcome := resolveActivatedAbilityRef(ref); outcome != abilityRefMatched {
		return nil
	}
	return &ref
}

// abilityRefOutcome is what restore found at a ref.
type abilityRefOutcome int

const (
	// abilityRefLost: no row, or more than one, carries the declared
	// label (Q3).
	abilityRefLost abilityRefOutcome = iota
	// abilityRefMatched: the row at the ref carries the declared label.
	abilityRefMatched
	// abilityRefMoved: the row at the ref does not, and exactly one
	// other row in the list does (Q2's fallback).
	abilityRefMoved
)

// resolveActivatedAbilityRef looks a ref up in the running catalog,
// with the owner's Q2 name check. It reads the catalog and nothing
// else, so capture, restore and the boot report ask one question.
func resolveActivatedAbilityRef(ref AbilityRef) (ActivatedAbilityShape, AbilityRef, abilityRefOutcome) {
	if ref.Slot != AbilitySlotActivated || ref.Key == "" || CatalogActivatedAbilities == nil {
		return ActivatedAbilityShape{}, ref, abilityRefLost
	}
	return resolveAbilityRow(ref, CatalogActivatedAbilities(ref.Key),
		func(a ActivatedAbilityShape) string { return a.Label })
}

// resolveTriggeredAbilityRef is resolveActivatedAbilityRef for the
// triggered slot (tier 4-2): the same lookup and the same Q2 check,
// over CatalogTriggers, with the row's Key as its declared name.
func resolveTriggeredAbilityRef(ref AbilityRef) (TriggeredAbility, AbilityRef, abilityRefOutcome) {
	if ref.Slot != AbilitySlotTriggered || ref.Key == "" || CatalogTriggers == nil {
		return TriggeredAbility{}, ref, abilityRefLost
	}
	return resolveAbilityRow(ref, CatalogTriggers(ref.Key),
		func(t TriggeredAbility) string { return t.Key })
}

// resolveAbilityRow is the one lookup behind both slots: the row at the
// ref if it carries the declared name, else the single row in the list
// that does (the ref rewritten to point at it), else lost (Q3).
func resolveAbilityRow[T any](ref AbilityRef, rows []T, name func(T) string) (T, AbilityRef, abilityRefOutcome) {
	var zero T
	if i, ok := abilityRefIndex(ref); ok && i < len(rows) && name(rows[i]) == ref.Name {
		return rows[i], ref, abilityRefMatched
	}
	if ref.Name == "" {
		return zero, ref, abilityRefLost
	}
	found := -1
	for j := range rows {
		if name(rows[j]) != ref.Name {
			continue
		}
		if found >= 0 {
			return zero, ref, abilityRefLost
		}
		found = j
	}
	if found < 0 {
		return zero, ref, abilityRefLost
	}
	moved := ref
	moved.Ref = withAbilityRefIndex(ref.Ref, found)
	return rows[found], moved, abilityRefMoved
}

// abilityRefLostInCatalog reports whether the running catalog has no
// row for a ref, in whichever slot it names — Q3's question, asked by
// the boot report as restore asks it.
func abilityRefLostInCatalog(ref AbilityRef) bool {
	switch ref.Slot {
	case AbilitySlotActivated:
		_, _, outcome := resolveActivatedAbilityRef(ref)
		return outcome == abilityRefLost
	case AbilitySlotTriggered:
		_, _, outcome := resolveTriggeredAbilityRef(ref)
		return outcome == abilityRefLost
	}
	return true
}

// abilityRefIndex parses the row index out of an own:<i> or
// grant:<bundle>:<i>:<n> ref. A grant ref must name the bundle its Key
// is, or the pair does not describe one row.
func abilityRefIndex(ref AbilityRef) (int, bool) {
	switch {
	case strings.HasPrefix(ref.Ref, abilityRefOwn):
		i, err := strconv.Atoi(strings.TrimPrefix(ref.Ref, abilityRefOwn))
		return i, err == nil && i >= 0
	case strings.HasPrefix(ref.Ref, abilityRefGrant):
		bundle, i, _, ok := parseGrantAbilityRef(ref.Ref)
		return i, ok && GrantKey(bundle) == ref.Key
	}
	return 0, false
}

// parseGrantAbilityRef splits grant:<bundle>:<i>:<n>. The bundle is
// everything between the prefix and the last two fields.
func parseGrantAbilityRef(r string) (bundle string, i, n int, ok bool) {
	rest := strings.TrimPrefix(r, abilityRefGrant)
	parts := strings.Split(rest, ":")
	if len(parts) < 3 {
		return "", 0, 0, false
	}
	var err1, err2 error
	i, err1 = strconv.Atoi(parts[len(parts)-2])
	n, err2 = strconv.Atoi(parts[len(parts)-1])
	bundle = strings.Join(parts[:len(parts)-2], ":")
	return bundle, i, n, err1 == nil && err2 == nil && i >= 0 && n >= 0 && bundle != ""
}

// withAbilityRefIndex rewrites the row index of a ref, keeping its
// kind and (for a grant) its bundle and instance.
func withAbilityRefIndex(r string, i int) string {
	if strings.HasPrefix(r, abilityRefGrant) {
		if bundle, _, n, ok := parseGrantAbilityRef(r); ok {
			return GrantedAbilityRef(bundle, i, n)
		}
	}
	return OwnAbilityRef(i)
}

// wellFormedAbilityRef is checkEffectKeys' test for a ref this binary
// can read at all: a known slot, a key, and a ref in the grammar.
func wellFormedAbilityRef(ref AbilityRef) bool {
	if !knownAbilitySlot(ref.Slot) || ref.Key == "" {
		return false
	}
	_, ok := abilityRefIndex(ref)
	return ok
}

// restoreCatalogAbility re-derives a stamped item from the running
// catalog: its Effect, target clause and mode clause from the row, and
// its ref rewritten when Q2's fallback found the row elsewhere. It
// reports false for Q3 — no row answers — and turns the item into a
// manual one: no body, no ref, no effect, nothing to re-check.
//
// A triggered row's clauses are restored only where the item held one
// (the snapshot's HasTargetSpec / HasModeSpec): a trigger's item takes
// its clause from the announcement, which may have had none to take. A
// clause built by TargetsFrom is re-derived after the whole game is
// back (rederiveTriggerClausesLocked), because it is handed the board.
func restoreCatalogAbility(out *StackItem, s *stackItemSnapshot) bool {
	lost := func() bool {
		out.Body = ""
		out.Params.Ability = nil
		out.Effect, out.targetSpec, out.modeSpec = nil, nil, nil
		return false
	}
	if out.Params.Ability == nil {
		// checkEffectKeys refuses a stamped item with no ref, so this
		// is a hand-built snapshot: treat it as lost rather than
		// resolving through nothing.
		return lost()
	}
	switch out.Body {
	case CatalogTriggeredBodyKey:
		row, ref, outcome := resolveTriggeredAbilityRef(*out.Params.Ability)
		if outcome == abilityRefLost {
			return lost()
		}
		out.Params.Ability = &ref
		out.Effect = row.Effect
		out.targetSpec, out.modeSpec = nil, nil
		if s.HasTargetSpec && row.TargetsFrom == nil {
			out.targetSpec = row.Targets
		}
		if s.HasModeSpec {
			out.modeSpec = row.Modes
		}
		return true
	default:
		row, ref, outcome := resolveActivatedAbilityRef(*out.Params.Ability)
		if outcome == abilityRefLost {
			return lost()
		}
		out.Params.Ability = &ref
		out.Effect = row.Effect
		out.targetSpec = row.Targets
		out.modeSpec = row.Modes
		return true
	}
}

// rederiveTriggerClauseLocked rebuilds the target clause of a restored
// triggered item whose row builds its clause from the trigger context
// (TriggeredAbility.TargetsFrom, #1223): called again with the item's
// carried Trigger and its source, exactly as the harvest called it.
// Only a row that does not read the board is ever stamped
// (TargetsFromReadsBoard), so the answer is the announcement's.
//
// The source is the card as the restored game has it — its instance ID
// survives every zone change — or, when it is in no zone (a token that
// has ceased to exist), a stand-in carrying the ID and controller the
// item recorded. Caller owns g.
func (g *Game) rederiveTriggerClauseLocked(it *StackItem, s *stackItemSnapshot) {
	if it == nil || s == nil || !s.HasTargetSpec || it.Body != CatalogTriggeredBodyKey || it.Params.Ability == nil {
		return
	}
	row, _, outcome := resolveTriggeredAbilityRef(*it.Params.Ability)
	if outcome == abilityRefLost || row.TargetsFrom == nil {
		return
	}
	var tc TriggerContext
	if it.Trigger != nil {
		tc = *cloneTriggerContext(it.Trigger)
	}
	source := Card{InstanceID: it.SourceCardID, Controller: it.Controller, Owner: it.Owner}
	if c := g.findCardByIDLocked(it.SourceCardID); c != nil {
		source = *c
	}
	it.targetSpec = row.TargetsFrom(tc, &source, g)
}

// LostStackAbility is one ability on a restore point's stack that this
// binary's catalog no longer has (Q3): the table is restored with it as
// a manual item, the source card flagged AbilitiesLostOnRestore, and
// the boot path logs it at ERROR.
type LostStackAbility struct {
	ItemID uuid.UUID
	// CardID and Card are the source card, as the snapshot names it.
	// Card is empty when the source is in no zone (a token that has
	// ceased to exist).
	CardID uuid.UUID
	Card   string
	Label  string
	Ref    AbilityRef
}

// LostStackAbilities lists every stamped stack item whose row this
// binary cannot find. It reads the snapshot and the running catalog
// and nothing else — the same lookup restore makes — so it gives the
// same answer before or after Restore.
func (s *GameSnapshot) LostStackAbilities() []LostStackAbility {
	var out []LostStackAbility
	for _, list := range [][]stackItemSnapshot{s.StackMeta, s.PendingTriggers} {
		for _, it := range list {
			if _, ok := catalogBodySlot(it.Body); !ok {
				continue
			}
			var ref AbilityRef
			if it.Params != nil && it.Params.Ability != nil {
				ref = *it.Params.Ability
				if !abilityRefLostInCatalog(ref) {
					continue
				}
			}
			out = append(out, LostStackAbility{
				ItemID: it.ID,
				CardID: it.SourceCardID,
				Card:   s.cardNameOf(it.SourceCardID),
				Label:  it.Label,
				Ref:    ref,
			})
		}
	}
	return out
}

// cardNameOf finds a card's name anywhere in the snapshot's zones.
func (s *GameSnapshot) cardNameOf(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	find := func(z *zoneSnapshot) (string, bool) {
		if z == nil {
			return "", false
		}
		for i := range z.Cards {
			if z.Cards[i].InstanceID == id {
				return z.Cards[i].Name, true
			}
		}
		return "", false
	}
	for _, z := range []*zoneSnapshot{s.Battlefield, s.Stack, s.Exile, s.PhasedOut} {
		if n, ok := find(z); ok {
			return n
		}
	}
	for i := range s.Seats {
		p := &s.Seats[i]
		for _, z := range []*zoneSnapshot{p.Library, p.Hand, p.Graveyard, p.Command, p.Emblems} {
			if n, ok := find(z); ok {
				return n
			}
		}
	}
	return ""
}

// flagAbilitiesLostLocked marks the source of an ability restore could
// not rebuild (Q3), wherever the card is. Caller holds g.mu or owns g.
func (g *Game) flagAbilitiesLostLocked(cardID uuid.UUID) {
	if c := g.findCardByIDLocked(cardID); c != nil {
		c.AbilitiesLostOnRestore = true
	}
}

// builds reports whether the declaration can put an item on the stack
// at all: a Build, or a declared Effect the engine builds from.
func (t TriggeredAbility) builds() bool { return t.Build != nil || t.Effect != nil }

// buildTriggerItemLocked makes the stack item for a triggered ability
// whose prompts (if any) are answered — the one place the harvest turns
// a declaration into an item (ADR 0041 P9, tier 4-2).
//
//   - A declared Effect and no Build: the engine builds it,
//     NewTriggeredItem(source, Key, Effect).
//   - A declared Effect and a Build: Build fills in what the engine
//     cannot know (a controller, a label, Params) and must leave
//     item.Effect nil; the engine installs the row's Effect. A Build
//     that set it anyway is an effectKeyFault — a panic in a test
//     binary; in production the closure is kept, unstamped, and the
//     census counts it.
//   - A Build and no Effect: the legacy shape, exactly as before.
//
// A declared item is then stamped with its catalog row
// (triggeredAbilityRefFor) unless Build already named a body of its
// own. Returns nil when Build suppressed the trigger.
//
// Caller must hold g.mu in write mode.
func (g *Game) buildTriggerItemLocked(t TriggeredAbility, ev Event, source Card, lki Characteristic) *StackItem {
	var item *StackItem
	switch {
	case t.Build != nil:
		item = t.Build(ev, &source, lki, g)
		// A Build that names a body of its own made a keyed item —
		// data already (tier 2's body keys, 4-0's engine triggers such
		// as suspend, madness and face-down ward) — and it wins over a
		// declared Effect beside it.
		if item == nil || t.Effect == nil || item.Body != "" {
			return item
		}
		if item.Effect != nil {
			effectKeyFault(fmt.Sprintf("game: trigger %q declares Effect and its Build set item.Effect as well — a fill-in Build must leave it nil (ADR 0041 P9)",
				labelOr(t.Key, item.Label)))
			return item
		}
		item.Effect = t.Effect
	case t.Effect != nil:
		item = NewTriggeredItem(&source, t.Key, t.Effect)
	default:
		return nil
	}
	if item.Body == "" {
		if ref := triggeredAbilityRefFor(t); ref != nil {
			item.Body = CatalogTriggeredBodyKey
			item.Params.Ability = ref
		}
	}
	return item
}
