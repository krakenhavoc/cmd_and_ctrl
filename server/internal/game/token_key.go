package game

import "strings"

// token_key.go gives a TOKEN a catalog identity (#521).
//
// Every other object in this engine answers "what do I do?" through
// the catalog, keyed by oracle ID: CatalogKey turns a Card into a
// key, catalogDef turns the key into a *CardDef, and the ability
// readers project the slot they want off it. A token has no printing
// and therefore no oracle ID, so before this it could not take that
// route at all — its abilities had to ride on the instance as Go
// closures (Card.ManaAbilities, Card.ActivatedAbilities), and three
// separate things went wrong downstream:
//
//   - the snapshot could not serialise a closure, so a board holding
//     a Treasure was censused as un-restorable and NO restore point
//     was written for as long as the Treasure sat there (ADR 0041
//     named this cost; it is not a rare edge — Treasure, Food, Clue,
//     Blood, Gold and Eldrazi Spawn are ordinary catalog output);
//   - the wire projection skipped a card with no oracle ID, so every
//     token's activated abilities were dropped on the way to the
//     client and a Food could not be cracked at all;
//   - and any future reader keying on oracle ID would have had to
//     grow its own token arm.
//
// The fix is the one emblems (EmblemKeyPrefix) and granted ability
// bundles (GrantKeyPrefix) already use: a SYNTHETIC key in a
// namespace of its own, filed in the same def map cards are filed in,
// resolved by the same CatalogLookup. "token:treasure" is a catalog
// entry like any other; nothing downstream needs to know a token is
// what it belongs to.
//
// Why it cannot collide with a real card: an oracle ID is a Scryfall
// UUID, which is 36 hex-and-dash characters and never contains a
// colon. This is the same guarantee EmblemKeyPrefix and
// GrantKeyPrefix rest on.
//
// And why the key is a FALLBACK rather than an override in
// CatalogKey: CR 707.2 makes a token COPY carry the copied card's
// copiable values, oracle ID included, so a token that is a copy of
// Llanowar Elves must keep resolving to Llanowar Elves' printed
// abilities. A real oracle ID therefore always wins; the token key
// answers only for an object that has none.

// TokenKeyPrefix marks a catalog key as a TOKEN TEMPLATE's rather
// than a card's. See the file comment for why it cannot collide with
// an oracle ID.
const TokenKeyPrefix = "token:"

// TokenKey is the catalog key a token template's printed abilities
// are registered under. Idempotent: a name that already carries the
// prefix is returned unchanged, so a caller may pass either spelling.
//
// The empty name returns the empty key — "no catalog entry" — rather
// than a bare prefix that would answer for every unnamed token at
// once.
func TokenKey(name string) string {
	if name == "" {
		return ""
	}
	if strings.HasPrefix(name, TokenKeyPrefix) {
		return name
	}
	return TokenKeyPrefix + name
}

// IsTokenKey reports whether a catalog key names a token template.
//
// Used where a key's NAMESPACE is the question rather than its
// contents — the snapshot's census, and any diagnostic that wants to
// tell a token's entry from a card's.
func IsTokenKey(key string) bool { return strings.HasPrefix(key, TokenKeyPrefix) }
