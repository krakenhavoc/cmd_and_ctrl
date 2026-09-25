package game

import (
	"strings"
	"sync"
)

// copy_grants.go — CR 707.9a and CR 707.9b: a copy effect's "except"
// clause may GRANT an ability or ADD a type, and what it grants is
// part of the copy's COPIABLE VALUES.
//
// # The rule, and why it needed a seam of its own
//
// CR 707.9a: an effect that copies an object "except" it gains an
// ability makes that ability part of the copiable values. The second
// half is the whole problem. A grant is not a continuous effect
// sitting on one permanent; it becomes part of what a LATER copy
// copies. A Clone that copies a Phantasmal Image is itself an
// Illusion with the sacrifice trigger.
//
// So a closure hanging off the replacement that made the copy is the
// wrong shape twice over: it cannot be copied again, and it cannot
// survive a snapshot. The grant has to ride in `PrintedValues` with
// the name, the type line and the oracle ID — pure data, copied by
// `CopiableValuesOf` for free, carried by the snapshot for free.
//
// # Where a granted ability LIVES
//
// Every ability the engine reads comes out of the catalog through
// `catalogDef`, keyed by `CatalogKey(c)` — which after a copy is the
// COPIED card's key, so a granted ability has nowhere of its own.
// ADR 0043 named that as the reason the clause was out of scope.
//
// The answer is the key itself. `CatalogKey` is already the one
// place a Card becomes a catalog answer (#940 put CR 708.2a's
// face-down silence there, as a "" return), so a card carrying
// grants returns a COMPOSITE key:
//
//	no grants  →  "<oracle_id>"                       (unchanged)
//	grants     →  "<oracle_id>|grant:<name>|grant:…"
//
// and `catalogDef` answers a composite key with the base card's
// definition MERGED with each grant's. Every existing reader — the
// trigger harvester, the layer pass's static gather, the activation
// path, the replacement gather, the view's auto bit — goes through
// those two functions and therefore sees a granted ability with no
// change of its own. That is the single lookup seam the design asked
// for; there is no second path to keep in sync.
//
// A grant's own definition is registered by the card that GRANTS it
// (`effects.Spec.Grants`), filed in the same `defs` map cards use
// under `GrantKey(name)`, exactly as #623 files an emblem's
// definition under `EmblemKey`. It is static catalog data, so the
// copy only ever has to carry its NAME — which is what makes the
// grant serialisable and copiable at once.
//
// # What a grant does NOT change
//
// A granted ability is an ability of the object like any other, so
// CR 613.1f ability removal takes it away with the rest:
// `CatalogAbilityKey`'s OWN half is "" for a permanent whose abilities
// were removed, copy grants included — only a later LAYER-6 grant (ADR
// 0093, granted_abilities.go) survives, because it is not the object's
// own. And a FACE-DOWN
// permanent has no text at all (CR 708.2a), so `CatalogKey`'s empty
// return wins over the grants too — a face-down copy is a 2/2 with
// nothing, granted ability included.

// GrantKeyPrefix marks a catalog key as a granted ability bundle's
// rather than a card's. Cards are keyed by oracle ID, which is a
// UUID and can never collide with it — the same guarantee
// EmblemKeyPrefix relies on.
const GrantKeyPrefix = "grant:"

// grantKeySeparator joins a card's catalog key to the grants riding
// on it. "|" is not a character any oracle ID, face suffix or grant
// name contains, so the split is unambiguous and BaseCatalogKey can
// recover the card identity from a composite key.
const grantKeySeparator = "|"

// GrantKey is the catalog key an ability grant is registered under.
// Idempotent: a name that already carries the prefix is returned
// unchanged, so a card file may declare either spelling.
func GrantKey(name string) string {
	if name == "" {
		return ""
	}
	if strings.HasPrefix(name, GrantKeyPrefix) {
		return name
	}
	return GrantKeyPrefix + name
}

// BaseCatalogKey strips the granted-ability suffix from a catalog
// key, leaving the CARD identity — the bare oracle ID, or the
// "<oracle_id>#N" of a non-front face.
//
// Use it wherever a catalog key is being used as an identity rather
// than as a lookup: deriving a second key from it (EmblemKey), or
// looking the card up in the registry of Specs rather than in the
// engine-facing def map. Every ordinary lookup wants the whole key,
// because the whole key is what carries the grants.
func BaseCatalogKey(key string) string {
	if i := strings.IndexByte(key, grantKeySeparator[0]); i >= 0 {
		return key[:i]
	}
	return key
}

// catalogKeyWithGrants appends a card's granted-ability keys to its
// base catalog key. An empty base (no oracle ID, or a face-down
// permanent's CR 708.2a silence) stays empty: an object with no text
// has no granted text either.
func catalogKeyWithGrants(base string, grants []string) string {
	if base == "" || len(grants) == 0 {
		return base
	}
	var b strings.Builder
	b.Grow(len(base) + len(grants)*(len(GrantKeyPrefix)+16))
	b.WriteString(base)
	for _, g := range grants {
		if g == "" {
			continue
		}
		b.WriteString(grantKeySeparator)
		b.WriteString(g)
	}
	return b.String()
}

// splitCatalogKeyGrants takes a composite key apart. The second
// return is nil for every key that is just a card, which is the fast
// path catalogDef checks before doing anything else.
func splitCatalogKeyGrants(key string) (base string, grants []string) {
	i := strings.IndexByte(key, grantKeySeparator[0])
	if i < 0 {
		return key, nil
	}
	return key[:i], strings.Split(key[i+1:], grantKeySeparator)
}

// mergedCatalogDef answers a composite key: the card's definition
// with every granted bundle's abilities appended.
//
// Returns nil only when NOTHING resolved — neither the card nor any
// grant — so `IsCatalogCard` and friends still read "no catalog
// entry" correctly. A grant on a card with no catalog entry of its
// own is the normal case, not an edge: a Phantasmal Image copying a
// vanilla imported bear has the sacrifice trigger and nothing else.
//
// The base may be EMPTY since ADR 0093: a layer-6 grant on an object
// whose own abilities are gone, a face-down 2/2, or an uncatalogued
// imported creature is the composite "|grant:<a>", and the answer is
// the bundles alone.
//
// Memoised, since ADR 0093 made composite keys common — Cryptolith Rite
// and twenty creatures is twenty composite lookups per trigger harvest
// where #665 had a handful of copies in the rarest game. The memo is
// VALIDATED rather than trusted: an entry records the exact *CardDef
// each part resolved to when it was built, and a hit re-resolves the
// parts (map reads) and compares the pointers before handing the
// merged def back. So a test that swaps CatalogLookup — the objection
// that kept #665 from caching at all — simply misses and re-merges,
// and the one allocation it saves is the merge itself. See
// mergedDefMemo.
func mergedCatalogDef(key string) *CardDef {
	if d, ok := mergedDefMemo.get(key); ok {
		return d
	}
	base, grants := splitCatalogKeyGrants(key)
	parts := make([]*CardDef, 0, 1+len(grants))
	parts = append(parts, lookupPart(base))
	for _, g := range grants {
		parts = append(parts, lookupPart(g))
	}
	d := mergeCatalogParts(parts)
	mergedDefMemo.put(key, base, grants, parts, d)
	return d
}

// lookupPart resolves one part of a composite key; the empty base of a
// grant-only composite resolves to nothing.
func lookupPart(key string) *CardDef {
	if key == "" {
		return nil
	}
	return CatalogLookup(key)
}

// mergeCatalogParts merges the resolved parts of a composite key: the
// base definition (parts[0], possibly nil) with every bundle appended.
// Nil when nothing resolved.
func mergeCatalogParts(parts []*CardDef) *CardDef {
	found := false
	for _, p := range parts {
		if p != nil {
			found = true
			break
		}
	}
	if !found {
		return nil
	}
	var merged CardDef
	if parts[0] != nil {
		merged = *parts[0]
	}
	for _, g := range parts[1:] {
		if g == nil {
			continue
		}
		merged.Triggered = concatTriggered(merged.Triggered, g.Triggered)
		merged.Static = concatStatic(merged.Static, g.Static)
		merged.Activated = concatActivated(merged.Activated, g.Activated)
		// ADR 0093 Decision 2 §2: a layer-6 grant may carry a MANA
		// ability (Cryptolith Rite), which #665's copy grants never
		// did. Merged here for the composite lookups that go through
		// the catalog whole; the mana ACCESSOR reads a grant's mana
		// abilities per bundle instead, so it can tell each row's
		// origin (manaAbilityRows).
		merged.ManaAbilities = concatMana(merged.ManaAbilities, g.ManaAbilities)
	}
	// ADR 0041 P9 (tier 4-2): a bundle granted twice contributes its
	// triggered rows twice, and each copy is a different instance of
	// the ability — the <n> of its ref.
	merged.Triggered = numberTriggerRowOccurrences(merged.Triggered)
	return &merged
}

// mergedDefMemoLimit bounds the memo. The number of distinct composite
// keys in a process is the number of distinct (card, grant set) pairs
// ever seen, which is small; the bound is a backstop against a
// pathological test, and hitting it clears the memo rather than
// evicting, because a re-merge is cheap and correct.
const mergedDefMemoLimit = 1024

type mergedDefEntry struct {
	base   string
	grants []string
	parts  []*CardDef
	def    *CardDef
}

// mergedDefMemo is mergedCatalogDef's validated memo. Process-wide and
// guarded by its own mutex: catalogDef runs under many games' read
// locks at once. The merged *CardDef it hands out is shared and must be
// treated as read-only, which is the contract every CardDef already has
// (the catalog's own are shared by every card with that oracle ID).
var mergedDefMemo = &mergedDefCache{}

type mergedDefCache struct {
	mu sync.RWMutex
	m  map[string]mergedDefEntry
}

// get answers from the memo only when every part still resolves to the
// very definition the entry was merged from. A hit allocates nothing.
func (c *mergedDefCache) get(key string) (*CardDef, bool) {
	c.mu.RLock()
	e, ok := c.m[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if lookupPart(e.base) != e.parts[0] {
		return nil, false
	}
	for i, g := range e.grants {
		if lookupPart(g) != e.parts[i+1] {
			return nil, false
		}
	}
	return e.def, true
}

func (c *mergedDefCache) put(key, base string, grants []string, parts []*CardDef, d *CardDef) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.m == nil || len(c.m) >= mergedDefMemoLimit {
		c.m = make(map[string]mergedDefEntry, 64)
	}
	c.m[key] = mergedDefEntry{base: base, grants: grants, parts: parts, def: d}
}

// The three concat helpers always allocate when there is something
// to add. Appending in place would be a bug with a long fuse: the
// base *CardDef is shared by every card in the game with that oracle
// ID, and an append into spare capacity would write a granted
// trigger into the catalog's own slice.

func concatTriggered(base, extra []TriggeredAbility) []TriggeredAbility {
	if len(extra) == 0 {
		return base
	}
	out := make([]TriggeredAbility, 0, len(base)+len(extra))
	out = append(out, base...)
	return append(out, extra...)
}

func concatStatic(base, extra []StaticAbility) []StaticAbility {
	if len(extra) == 0 {
		return base
	}
	out := make([]StaticAbility, 0, len(base)+len(extra))
	out = append(out, base...)
	return append(out, extra...)
}

func concatMana(base, extra []ManaAbilityShape) []ManaAbilityShape {
	if len(extra) == 0 {
		return base
	}
	out := make([]ManaAbilityShape, 0, len(base)+len(extra))
	out = append(out, base...)
	return append(out, extra...)
}

func concatActivated(base, extra []ActivatedAbilityShape) []ActivatedAbilityShape {
	if len(extra) == 0 {
		return base
	}
	out := make([]ActivatedAbilityShape, 0, len(base)+len(extra))
	out = append(out, base...)
	return append(out, extra...)
}
