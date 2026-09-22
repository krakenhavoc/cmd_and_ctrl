package game

import "github.com/google/uuid"

// mana_source.go — #1212: WHAT a mana came from, snapshotted at the
// moment it was made.
//
// `ManaToken.Source` is a uuid, and for the commonest reader of a
// source it is useless by the time anyone asks. A Treasure is
// sacrificed as part of its own mana ability's cost, so by the time
// "if mana from a Treasure was spent to cast it" resolves the
// permanent is in a graveyard — and a Treasure TOKEN has ceased to
// exist entirely (CR 111.7), which is the same failure #596 hit with
// sacrificed Food and #761 hit with the mana itself. A uuid that
// names nothing is not a record.
//
// So the token carries a SNAPSHOT of the kinds that matter to printed
// text, taken where the mana is minted and before the cost that
// removes the source can run. Six bits, closed, named after the words
// the cards use:
//
//	"if mana from a Treasure was spent"          Hired Hexblade
//	"three or more mana from creatures"          Inga and Esika
//	"mana from an artifact was spent"            Shadow the Hedgehog
//	"{S}" in a cost                              the snow half, unused
//
// NOT the full type line, and not the Card. A copy of the card would
// be a second object the snapshot, the clone and the undo stack all
// have to carry, and it would go stale in a different way — the
// question is never "what is that permanent now", it is "what was it
// when it made this mana". A bitset answers exactly that and costs
// two bytes.

// ManaSourceKinds is the closed set of facts about a mana's source
// that printed text asks about, snapshotted when the mana was made.
//
// A bitset rather than a list of type strings: the set is closed by
// construction (a new kind is a new constant here and a new reader on
// ManaSpent), it rides ManaToken by value into the clone and the
// snapshot, and an unknown bit cannot appear — which is what stops a
// reader guessing.
type ManaSourceKinds uint16

const (
	// ManaSourceSnow is the SNOW supertype (CR 205.4h) on the
	// permanent that produced the mana — what a `{S}` symbol in a
	// cost would be paid with (CR 107.4g).
	//
	// Recorded and read by nothing in the catalog today: a scan of
	// the Scryfall dump finds no card whose oracle text asks whether
	// "snow mana was spent", because snow reaches a payment as a COST
	// symbol and `ParsedCost.HasSnow` is still informational
	// (mana_cost.go:39). It is here because it is one bit of the same
	// snapshot and because the `{S}` work has nowhere else to read
	// from; stated rather than left as a silent hook.
	ManaSourceSnow ManaSourceKinds = 1 << iota

	// ManaSourceTreasure is the TREASURE subtype — the kind eighteen
	// printed cards ask about, and the only one whose source is
	// reliably gone by the time it is asked.
	ManaSourceTreasure

	// ManaSourceCreature is "mana from creatures" (Inga and Esika).
	// The card type at mint time, so a Dryad Arbor's mana is from a
	// creature and a Birds of Paradise animated by nothing is too.
	ManaSourceCreature

	// ManaSourceLand is the land half of the same question. No card
	// in the catalog reads it yet; it is the cheapest bit in the set
	// and the one a "mana from a land" clause would want.
	ManaSourceLand

	// ManaSourceArtifact is "mana from an artifact" (Shadow the
	// Hedgehog). A Treasure sets this as well as ManaSourceTreasure —
	// a Treasure IS an artifact, and a reader that had to know the
	// subtype implies the type would be a reader that gets it wrong.
	ManaSourceArtifact

	// ManaSourceEnchantment completes the permanent types a mana
	// ability is printed on (Wild Growth's host is the land, but
	// Nykthos and Serra's Sanctum are enchantment-adjacent shapes and
	// a "mana from an enchantment" clause is one printing away).
	ManaSourceEnchantment
)

// Has reports whether every bit in `want` is set. The zero `want` is
// "no wish", and answers true — a caller with nothing to ask for is
// satisfied by any source.
func (k ManaSourceKinds) Has(want ManaSourceKinds) bool {
	return k&want == want
}

// HasAny reports whether ANY bit in `want` is set. The reader a wish
// uses: "prefer a source that is a Treasure OR snow" is one call.
// False for the zero `want`, which is the opposite default from Has
// and the right one — an empty wish is satisfied by nothing in
// particular, so nothing is preferred.
func (k ManaSourceKinds) HasAny(want ManaSourceKinds) bool {
	return k&want != 0
}

// manaSourceKindsOf snapshots the kinds of one permanent. THE single
// decider: every mint site calls this and nothing else classifies a
// mana source, the way restrictionsFor is the single decider for a
// token's restrictions (mana_restriction.go:234).
//
// Effective characteristics, not printed: a land animated into a
// creature really is producing mana from a creature while it is one
// (CR 613 layer 4), and a Mycosynth Lattice makes every source an
// artifact source. That is the same read `ManaSpendForCast` makes of
// the object a restricted token is tested against.
func manaSourceKindsOf(c Card) ManaSourceKinds {
	var k ManaSourceKinds
	if c.HasSupertype("snow") {
		k |= ManaSourceSnow
	}
	if c.HasSubtype("Treasure") {
		k |= ManaSourceTreasure
	}
	if c.IsCreature() {
		k |= ManaSourceCreature
	}
	if c.IsLand() {
		k |= ManaSourceLand
	}
	if c.IsArtifact() {
		k |= ManaSourceArtifact
	}
	if c.IsEnchantment() {
		k |= ManaSourceEnchantment
	}
	return k
}

// manaSourceKindsLocked is manaSourceKindsOf for a caller that has an
// ID rather than a card — the mint sites that queue a colour pick and
// the effect-driven "add {B}{B}{B}" path, where the source may be a
// spell that is not on the battlefield at all.
//
// The battlefield and nowhere else. A source that is not a permanent
// has no kinds worth recording: "mana from a Treasure" is a question
// about a permanent's ability (CR 605.1a), and a resolving Dark
// Ritual is not one. Zero is the honest answer and the weaker one.
//
// Caller must hold g.mu.
func (g *Game) manaSourceKindsLocked(sourceID uuid.UUID) ManaSourceKinds {
	if sourceID == uuid.Nil || g.Battlefield == nil {
		return 0
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == sourceID {
			return manaSourceKindsOf(g.Battlefield.Cards[i])
		}
	}
	return 0
}

// CatalogWantsManaFrom is the catalog hook for #1212's auto-tapper
// ordering hint: the kinds a card would RATHER be paid with, because
// its printed text reads them back ("if mana from a Treasure was
// spent to cast it"). Nil (no catalog) means no card wishes for
// anything.
//
// A DECLARATION on the spec, for the reason WantsDistinctColors is
// one (ADR 0068 §5): the alternative is a scan of the oracle text
// that quietly stops matching when a card words the clause
// differently.
var CatalogWantsManaFrom func(oracleID string) ManaSourceKinds

// WantedManaSourcesFor is the kinds this card's own text reads back,
// or zero. The planner's only question, and the /autotap preview
// endpoint's — the preview has to plan the same payment the cast
// would, or the two routes to the same spell spend different mana.
func WantedManaSourcesFor(card Card) ManaSourceKinds {
	if CatalogWantsManaFrom == nil {
		return 0
	}
	return CatalogWantsManaFrom(CatalogKey(card))
}
