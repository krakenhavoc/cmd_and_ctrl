package game

// x_bound.go — the engine's own bound on an announced X (#3895).
//
// CR 107.3a lets the caster choose X, and CR 107.1b rules out only a
// negative number, so the rules put no top on it. The engine does:
// cost handling multiplies the announced X into the generic demand,
// folds a "spend only … on X" restriction into one requirement per
// mana (spend_only.go), and sizes life, loyalty, energy, blight,
// discard and tap payments by it. Every one of those reads X before
// the payment is known to be possible, so X is bounded where it
// enters, before any of them runs:
//
//   - castSpellLocked, for every spell (and every parked, assisted or
//     controlled cast, which all re-enter it);
//   - activateCatalogAbilityLocked, for every catalog activation, with
//     its X-sized costs (loyalty −X, pay X {E}, discard, reveal,
//     sacrifice or tap X);
//   - the sandbox ActivateAbility and AnnounceTrigger verbs, which
//     record an X on a stack item;
//   - priceCastLocked, the pricer the auto-tap preview and the
//     legal-move enumerator ask;
//   - the auto-tap preview's ?x= query (internal/lobby), the
//     enumerator's Options.MaxX and open X range (internal/legal) and
//     the MCP seat's X value (internal/mcpseat).
//
// Refused with ErrInvalidParam, the answer the printed ceiling
// (SpellXCeilingLocked) gives, and for its reason: a client that
// announces an X the engine will not take is wrong about the move it
// is making, and quietly casting a different spell would hide that.
//
// # Why a fixed number, and why this one
//
// Not "the most the caster could pay". Permissive mode casts a spell
// whose mana was paid on paper (ADR 0068), and ForceCast overrides the
// mana gate, so what the pool and the untapped sources could produce
// is not a bound the engine may put on the announcement; strict mode's
// mana gate already refuses an X the caster cannot pay.
//
// One million, the number PayAmountNoMaxCeiling already sets for a
// chosen amount with no printed ceiling, because the two are the same
// question. It is far past any finite X a game needs — every life
// total, library and board is orders of magnitude smaller — and still
// payable: the run-length mana pool (#3698) holds a million mana from
// an infinite-mana combo as a handful of runs, so a table that made
// that much can spend it. And it keeps the arithmetic safe: no printed
// cost has more than three {X}, so XSlots*X stays far inside an int.
//
// # Below the bound
//
// A plain {X} is a count the run-length pool pays in O(runs). A
// restricted X ("spend only black mana on X") is folded into X
// identical requirements (foldSpendOnly), kept as one counted run
// (#3909), so the fold is constant in X too. The pool solver reads
// the run as it is (#3904): its coloured pass books a run of
// requirements against a run of mana in one step (coloured_pass.go),
// and what it leaves unpaid comes back as runs, so MissingFor, the paper-payment picker and its declaration
// name a long shortfall once, with its count (maxListedRun). XMana and
// GenericWithX saturate rather than wrap, and no list reserves capacity
// by X.
//
// The other numbers a client sends are bounded where they are read:
// the Phyrexian life count sizes nothing (PhyrexianLifePlan), a paper
// declaration's counts are checked against the shortfall as they are
// summed (paperShortfall.tokensFor), and the auto-tapper's plan is
// sized by the sources, not the requirements.
const MaxAnnouncedX = 1_000_000

// announcedXInRange reports whether x is an X the engine accepts at
// an announcement: 0 through MaxAnnouncedX.
func announcedXInRange(x int) bool {
	return x >= 0 && x <= MaxAnnouncedX
}

// maxManaDemand is where XMana and GenericWithX saturate: a demand
// larger than any pool can hold, small enough that the sums callers
// make of it cannot wrap. Every announcement is bounded by
// MaxAnnouncedX first, so this is a second line rather than a rule.
const maxManaDemand = 1 << 30

// satAddMana is a+b for two non-negative mana counts, saturating at
// maxManaDemand.
func satAddMana(a, b int) int {
	if a >= maxManaDemand || b >= maxManaDemand || a+b >= maxManaDemand {
		return maxManaDemand
	}
	return a + b
}

// satMulMana is a*b for two non-negative counts, saturating at
// maxManaDemand.
func satMulMana(a, b int) int {
	if a <= 0 || b <= 0 {
		return 0
	}
	if a >= maxManaDemand || b > maxManaDemand/a {
		return maxManaDemand
	}
	return min(a*b, maxManaDemand)
}

// maxFoldedMana is the most restricted mana foldSpendOnly folds into
// a run: three {X} at MaxAnnouncedX plus room for any generic part.
// Past it the fold writes one requirement no mana can pay instead (see
// foldSpendOnly). A run is one entry whatever its length, but the
// readers that write it out (ParsedCost.unfolded) are sized by it.
const maxFoldedMana = 4 * MaxAnnouncedX

// maxPreallocSlots is the most capacity a per-mana list (a paper
// payment's slots, a missing-mana breakdown) reserves up front. A run
// longer than maxListedRun is one entry in such a list (#3904), so it
// rarely grows past this.
const maxPreallocSlots = 64

// maxListedGeneric is the longest generic shortfall MissingFor names
// one {1} at a time. A longer one is a single {N}, the way a printed
// cost writes it, so an X-sized shortfall is one symbol on the wire
// rather than X of them.
const maxListedGeneric = 16
