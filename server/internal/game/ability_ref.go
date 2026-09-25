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
// Triggered abilities are NOT stamped yet. Their Effect is still a
// closure a hand-written Build may have filled with trigger-time
// values; slice 4-2 makes it a declared row field. Until then this
// binary does not register "catalog/triggered" either, so it refuses a
// file from 4-2 rather than restoring a trigger it cannot rebuild.
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
	// Slot is which list of the definition: "activated". (4-2 adds
	// "triggered"; this binary refuses any slot it does not read.)
	Slot string `json:"slot"`
	// Ref is ADR 0093 Decision 5's row ref, counted in the full
	// declared list before designation gating.
	Ref string `json:"ref"`
	// Name is the row's declared label at announce, checked on restore
	// (the owner's answer to Q2).
	Name string `json:"name"`
}

// AbilitySlotActivated is the activated-ability list of a CardDef.
const AbilitySlotActivated = "activated"

// knownAbilitySlot reports whether this binary can resolve a ref in
// the slot.
func knownAbilitySlot(slot string) bool { return slot == AbilitySlotActivated }

// CatalogActivatedBodyKey is the body an activated ability's stack item
// names. It is the refusal token of P10: see the file comment.
const CatalogActivatedBodyKey = "catalog/activated"

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
	rows := CatalogActivatedAbilities(ref.Key)
	if i, ok := abilityRefIndex(ref); ok && i < len(rows) && rows[i].Label == ref.Name {
		return rows[i], ref, abilityRefMatched
	}
	if ref.Name == "" {
		return ActivatedAbilityShape{}, ref, abilityRefLost
	}
	found := -1
	for j := range rows {
		if rows[j].Label != ref.Name {
			continue
		}
		if found >= 0 {
			return ActivatedAbilityShape{}, ref, abilityRefLost
		}
		found = j
	}
	if found < 0 {
		return ActivatedAbilityShape{}, ref, abilityRefLost
	}
	moved := ref
	moved.Ref = withAbilityRefIndex(ref.Ref, found)
	return rows[found], moved, abilityRefMoved
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
func restoreCatalogAbility(out *StackItem) bool {
	if out.Params.Ability == nil {
		// checkEffectKeys refuses a stamped item with no ref, so this
		// is a hand-built snapshot: treat it as lost rather than
		// resolving through nothing.
		out.Body = ""
		return false
	}
	row, ref, outcome := resolveActivatedAbilityRef(*out.Params.Ability)
	if outcome == abilityRefLost {
		out.Body = ""
		out.Params.Ability = nil
		out.Effect, out.targetSpec, out.modeSpec = nil, nil, nil
		return false
	}
	out.Params.Ability = &ref
	out.Effect = row.Effect
	out.targetSpec = row.Targets
	out.modeSpec = row.Modes
	return true
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
			if it.Body != CatalogActivatedBodyKey {
				continue
			}
			var ref AbilityRef
			if it.Params != nil && it.Params.Ability != nil {
				ref = *it.Params.Ability
				if _, _, outcome := resolveActivatedAbilityRef(ref); outcome != abilityRefLost {
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
