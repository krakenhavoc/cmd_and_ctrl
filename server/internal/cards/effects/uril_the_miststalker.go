package effects

// Uril, the Miststalker — Legendary Creature — Beast, {2}{R}{G}{W}, 5/5:
//
//	"Hexproof"
//	"Uril, the Miststalker gets +2/+2 for each Aura attached to it."
//
// One of the two commanders #77 named by hand, and the honest half of
// it ships here: a five-mana 5/5 with hexproof that your opponents
// cannot remove by targeting.
//
// # DECLARED SIMPLIFICATION — the Aura clause is deferred, not skipped
//
// "+2/+2 for each Aura attached to it" is NOT implemented, and it is
// not implementable today: **there is no attachment relation in the
// engine**. `game.Card` has no `AttachedTo` field, an Aura resolves
// to the battlefield attached to nothing, and an Equipment is an
// artifact that sits there. The design exists — ADR 0036 specifies
// `Card.AttachedTo TargetRef`, one state-based action, one wire key
// and one rendering rule — but it is a spike with no code, scheduled
// for S33 (#280 / #76).
//
// The count would be a one-line predicate over that field the day it
// lands, so this card is written to need exactly that one line and
// nothing else: a layer 7c `Static` counting attached Auras. Until
// then Uril is a vanilla 5/5 with a real, enforced hexproof.
//
// # Why ship it at all
//
// Overrun's file set the bar: do not ship a card where EVERY word
// would be a declared no-op (which is why Heroic Intervention waited
// for S25). Uril clears it — hexproof is half the card's text and is
// fully enforced, and a 5/5 hexproof body is a functioning voltron
// commander in a format where it takes three connections to kill
// someone (CR 903.14a). Bruna, Light of Alabaster, the other
// commander #77 named, does NOT clear it: her entire printed value
// is a single attach-everything trigger, and a 5/5 flier with
// vigilance is what the deck importer already gives you without a
// catalog entry. Bruna waits for S33.
func init() {
	Register(Spec{
		OracleID:        "4308a020-48cf-45fe-8074-dd5d0ac6d12d",
		Name:            "Uril, the Miststalker",
		PrintedKeywords: []string{"hexproof"},
	})
}
