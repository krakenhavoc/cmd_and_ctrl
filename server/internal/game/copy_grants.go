package game

import "strings"

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
// `CatalogAbilityKey` returns "" for a permanent that has lost all
// abilities, before this file is ever consulted. And a FACE-DOWN
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
// Deliberately NOT cached. The merge allocates, but only for an
// object that actually carries a grant — a handful of permanents in
// the rarest game — and a process-lifetime cache would go stale the
// moment a test swapped CatalogLookup. The fast path, which is every
// other card in the game, never reaches here at all.
func mergedCatalogDef(base string, grants []string) *CardDef {
	found := false
	var merged CardDef
	if d := CatalogLookup(base); d != nil {
		merged = *d
		found = true
	}
	for _, key := range grants {
		g := CatalogLookup(key)
		if g == nil {
			continue
		}
		found = true
		merged.Triggered = concatTriggered(merged.Triggered, g.Triggered)
		merged.Static = concatStatic(merged.Static, g.Static)
		merged.Activated = concatActivated(merged.Activated, g.Activated)
	}
	if !found {
		return nil
	}
	return &merged
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

func concatActivated(base, extra []ActivatedAbilityShape) []ActivatedAbilityShape {
	if len(extra) == 0 {
		return base
	}
	out := make([]ActivatedAbilityShape, 0, len(base)+len(extra))
	out = append(out, base...)
	return append(out, extra...)
}
