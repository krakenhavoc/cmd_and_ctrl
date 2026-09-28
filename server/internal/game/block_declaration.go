package game

import "github.com/google/uuid"

// block_declaration.go holds the declare-blockers VERB: the set-based
// DeclareBlockers, the validator it runs before it stores anything,
// and the option generator the enumerator and the #328 auto-pass
// signal read (#750, ADR 0045 addendum Decisions 12-14).
//
// Why a set. CR 509.1 declares blockers as ONE turn-based action, and
// CR 509.1b's block COUNTS — menace's minimum of 2, Hungering Hydra's
// maximum of 1 — are properties of a whole declaration, not of a
// pair. A per-pair verb has no complete declaration to judge, which
// is why this engine used to accept a lone block on a menace attacker
// and silently undo it later: first at the damage step, then (#830)
// at the declaration's lock-in. Either way the defender had already
// been told the block was good.
//
// The rule now is Decision 13's: THE ENGINE NEVER HOLDS A BLOCK IT
// WOULD REFUSE. A declaration is validated as it will be AFTER the
// action, and if anything in it is illegal nothing is stored, no
// event is emitted, and the first refusal is returned with a reason
// the player can read. A two-creature menace block is legal only as a
// pair, so it must arrive as a pair.
//
// DeclareBlocker (mutations.go) is exactly a one-entry DeclareBlockers
// and keeps working for every ordinary block.
//
// The validator below is a list of set checks: the per-pair ones, the
// per-attacker count bounds, and — #1507, Decision 43 — the
// whole-combat BlockRule.Limit (Silent Arbiter's "no more than one
// creature can block each combat"), and — #1597 — the CR 509.1c
// blocking REQUIREMENTS of Decision 15 (block_requirements.go).

// BlockDeclaration is one (blocker, attacker) pairing in a block
// declaration. A declaration is a slice of them, applied as a unit.
type BlockDeclaration struct {
	Blocker  uuid.UUID
	Attacker uuid.UUID
}

// blockEntry is a resolved BlockDeclaration: the validator looks both
// cards up once and hands the pointers to the writer, so the store
// pass repeats no battlefield walks.
type blockEntry struct {
	blocker  *Card
	attacker *Card
}

// DeclareBlockers declares a whole set of blocks as one action
// (CR 509.1). Gated by the declare_blockers step.
//
// ALL OR NOTHING. Every entry is checked before any of them is
// stored: the blocker is a creature on the battlefield, the attacker
// is on the battlefield, the blocker's controller is the attack's
// defending player (CR 802.4a, #1339), the pair passes
// BlockPairRefusalLocked
// (CR 509.1b's restrictions and evasion keywords), and — once the
// whole proposed set is known — every attacker whose set of blockers
// this action CHANGES is within its block-count bounds
// (blockerBoundsLocked). If anything is refused nothing is stored and
// the first refusal is returned.
//
// This is deliberately unlike DeclareAttackers, which silently skips
// ineligible entries. A two-creature menace block is legal only as a
// pair, so skipping half of it would leave behind exactly the illegal
// declaration this verb exists to refuse.
//
// Errors: ErrGameNotActive, ErrWrongStep outside declare_blockers,
// ErrEmptyBlockerSet for an empty set, ErrCardNotFound when either
// card is missing from the battlefield, ErrNotACreature for a
// non-creature blocker, and a *BlockRefusedError (which wraps
// ErrIllegalBlock) for a refused pair, a refused count OR a broken
// whole-combat limit (declaration_limit, #1507). Idempotent:
// re-declaring a pairing that is already stored changes nothing and
// is not re-judged.
//
// Like DeclareBlocker, this verb only STAGES the pairings (#830).
// Nothing is announced until commitBlockDeclarationLocked locks the
// declaration in at the first priority boundary inside the step, so
// a defender may still revise it — and every revision goes through
// this same validator, which is what keeps the stored declaration
// legal at every moment, not just at the end.
//
// The attacker must be attacking the blocker's controller, a
// planeswalker they control or a battle they protect. The sandbox used
// to accept a "pre-emptive" block on a creature that was not attacking
// at all, and — the bug #1339 closed — a block on a creature attacking
// somebody else; both are refused with not_defending now, which is
// the answer the option generator has always given.
func (g *Game) DeclareBlockers(decls []BlockDeclaration) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.declareBlockersLocked(decls)
}

// declareBlockersLocked is DeclareBlockers without the lock. Caller
// must hold g.mu in write mode.
func (g *Game) declareBlockersLocked(decls []BlockDeclaration) error {
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if g.Turn.Step != StepDeclareBlockers {
		return ErrWrongStep
	}
	if len(decls) == 0 {
		return ErrEmptyBlockerSet
	}
	// Layers must be fresh so the pair check and HasKeyword read the
	// current effective characteristics — flying granted by an anthem
	// this turn, menace granted by an aura, a land Urborg made a
	// Swamp.
	g.RecomputeLayersIfStaleLocked()
	entries, after, err := g.checkBlockDeclarationAfterLocked(g.currentBlockAssignmentLocked(), decls)
	if err != nil {
		return err
	}
	if after == nil {
		// A re-declaration of pairings already stored: nothing moves.
		return nil
	}
	for _, e := range entries {
		// #1706: the blocker's whole set as the action leaves it — one
		// attacker for an ordinary blocker, every attacker it now blocks
		// for one that can block more.
		e.blocker.setBlockingSet(after[e.blocker.InstanceID])
		// Attacking and blocking are mutually exclusive per card.
		e.blocker.AttackingTarget = uuid.Nil
	}
	return nil
}

// currentBlockAssignmentLocked is the block declaration as it stands:
// blocker instance ID → the attackers it is pointed at (#1706: more
// than one for a creature that can block more). The validator's
// starting point, and the thing a proposed declaration is applied on
// top of.
//
// Caller must hold g.mu (read or write).
func (g *Game) currentBlockAssignmentLocked() blockAssignment {
	out := blockAssignment{}
	if g.Battlefield == nil {
		return out
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if atks := c.BlockedAttackers(); len(atks) > 0 {
			out[c.InstanceID] = atks
		}
	}
	return out
}

// checkBlockDeclarationLocked is Decision 13's validator: it decides
// whether `decls` applied on top of `base` would be a legal block
// declaration, and returns the resolved entries when it would be.
//
// PURE. It walks the battlefield and reads effective characteristics
// and nothing else — no card is written, no event emitted — so
// BlockOptionsLocked can run it over hypothetical sets and the
// enumerator's read-only contract (ADR 0045 §3) holds.
//
// `base` is the assignment the declaration starts from, normally
// currentBlockAssignmentLocked(). An entry naming a blocker that is
// already blocking RE-POINTS it when the blocker can block only one
// attacker, which is how the sandbox's "re-declare blocker" works; for
// a blocker that can block more (#1706, BlockCapacity) it ADDS the
// attacker, and is refused with blocker_capacity once there is no room
// left.
//
// What is checked, in order:
//
//   - per entry: both cards are on the battlefield, the blocker is a
//     creature, its controller is the attack's defending player
//     (CR 802.4a, blockDefenderRefusalLocked — skipped for a pairing
//     `base` already holds), and the pair passes
//     BlockPairRefusalLocked;
//   - per attacker whose blocker set this action CHANGES: the count
//     bounds (CR 509.1b). An attacker that LOSES a re-pointed blocker
//     is changed too, so a defender cannot pull one creature out of a
//     menace block and leave an illegal one behind;
//   - over the whole combat: every BlockRule.Limit on the battlefield
//     or in the turn-scoped registry (#1507), counted across every
//     block stored, whoever made it (blockLimitRefusalLocked).
//
// Attackers the action does not touch are NOT re-judged. An attacker
// that gains menace after a legal block keeps that block (CR 509.1b:
// restrictions are checked only as blockers are declared), and so
// does one whose second blocker has died (CR 509.1h, #715).
//
//   - and, last, the CR 509.1c requirements (#1597,
//     blockRequirementRefusalLocked): a declaration that makes a
//     requirement the defending player could still obey unobeyable is
//     refused, the one set check Decision 15 left for later.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) checkBlockDeclarationLocked(base blockAssignment, decls []BlockDeclaration) ([]blockEntry, error) {
	entries, _, err := g.checkBlockDeclarationAfterLocked(base, decls)
	return entries, err
}

// checkBlockDeclarationAfterLocked is checkBlockDeclarationLocked that
// also hands back the assignment the action would leave behind — nil
// when it changes nothing — which is what the verb writes.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) checkBlockDeclarationAfterLocked(base blockAssignment, decls []BlockDeclaration) ([]blockEntry, blockAssignment, error) {
	entries, after, err := g.checkBlockRestrictionsLocked(base, decls)
	if err != nil || after == nil {
		return entries, after, err
	}
	if e := g.blockRequirementRefusalLocked(base, after, decls, nil); e != nil {
		return nil, nil, e
	}
	return entries, after, nil
}

// checkBlockDeclarationCachedLocked is checkBlockDeclarationLocked with
// a memo of reach(base) per defending player, for a caller judging many
// declarations against one base (blockOptionsLocked). nil caches
// nothing.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) checkBlockDeclarationCachedLocked(base blockAssignment, decls []BlockDeclaration, cache blockReachCache) ([]blockEntry, error) {
	entries, after, err := g.checkBlockRestrictionsLocked(base, decls)
	if err != nil {
		return nil, err
	}
	if after == nil {
		// Nothing changes: a re-declaration of pairings already stored.
		return entries, nil
	}
	if e := g.blockRequirementRefusalLocked(base, after, decls, cache); e != nil {
		return nil, e
	}
	return entries, nil
}

// checkBlockRestrictionsLocked is every check of the validator but the
// requirements: the per-pair, count and limit checks. It returns the
// entries and the assignment the action would leave behind — nil when
// the action changes nothing. The requirement search reads it to judge
// its own witness, so the search never recurses into itself.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) checkBlockRestrictionsLocked(base blockAssignment, decls []BlockDeclaration) ([]blockEntry, blockAssignment, error) {
	entries := make([]blockEntry, 0, len(decls))
	// after is the assignment the action would leave behind.
	after := base.clone()
	// touched is every attacker whose blocker set changes, in the
	// order it was first touched, so the refusal a caller sees is
	// stable. A slice, not a set: a declaration is a handful of
	// entries, and the linear scan beats a map allocation.
	var touched []uuid.UUID
	mark := func(id uuid.UUID) {
		if id == uuid.Nil {
			return
		}
		for _, t := range touched {
			if t == id {
				return
			}
		}
		touched = append(touched, id)
	}
	for _, d := range decls {
		// The attacker is looked up first so a declaration naming a
		// missing attacker reports that, as the per-pair verb always
		// has.
		attacker := findBattlefieldCard(g, d.Attacker)
		if attacker == nil {
			return nil, nil, ErrCardNotFound
		}
		blocker := findBattlefieldCard(g, d.Blocker)
		if blocker == nil {
			return nil, nil, ErrCardNotFound
		}
		if !blocker.IsCreature() {
			return nil, nil, ErrNotACreature
		}
		// CR 802.4a / 509.1a (#1339): a defending player blocks only
		// creatures attacking THEM, a planeswalker they control or a
		// battle they protect. defendingPlayerForAttackerLocked is the
		// same resolution blockOptionsLocked uses to decide which
		// attackers a seat is offered, so the verb and the generator
		// give one answer (Decision 14) — including for an attacker
		// whose planeswalker or battle has left (CR 506.4c, #1364).
		//
		// A pairing already in `base` is not re-judged. CR 508.7a /
		// 509.1h: an attacker reselected onto another player after it
		// was blocked (#1343) stays blocked by the same creatures, and
		// a declaration that repeats that pairing — alone or beside a
		// new one — must not be refused for a block the rules say is
		// still standing.
		if !base.has(d.Blocker, d.Attacker) {
			if r := g.blockDefenderRefusalLocked(attacker, blocker); !r.Legal() {
				return nil, nil, g.blockRefusedErrorLocked(attacker, blocker, r)
			}
		}
		// CR 509.1b, per pair: restrictions, evasion keywords and the
		// block rules with a parameter (#750). BlockPairRefusalLocked
		// is the ONE pair check the enumerator and the #328 signal
		// also read, so the three can never disagree (ADR 0045 §3).
		//
		// The count minimum deliberately does NOT live there: that
		// gate runs per pair, so it would refuse the FIRST of two
		// menace blockers and make every menace attacker unblockable.
		// A count is judged below, where the whole set is known.
		if r := g.BlockPairRefusalLocked(attacker, blocker); !r.Legal() {
			return nil, nil, g.blockRefusedErrorLocked(attacker, blocker, r)
		}
		entries = append(entries, blockEntry{blocker: blocker, attacker: attacker})
		if after.has(d.Blocker, d.Attacker) {
			// Already pointed there: idempotent, and not re-judged.
			continue
		}
		// #1706, CR 509.1a/b: how many attackers this blocker may
		// block. One — the ordinary creature — RE-POINTS it, exactly
		// as the sandbox always has, and the attacker it leaves is
		// touched. More, and the attacker is ADDED while there is room;
		// a full blocker is refused, never silently re-pointed, because
		// dropping an attacker the defender chose would be a block
		// they did not ask to take back.
		switch capacity, cur := BlockCapacity(blocker), after[d.Blocker]; {
		case capacity == 1:
			for _, prev := range cur {
				mark(prev)
			}
			after[d.Blocker] = []uuid.UUID{d.Attacker}
		case capacity == 0 || len(cur) < capacity:
			after[d.Blocker] = append(cur, d.Attacker)
		default:
			return nil, nil, g.blockRefusedErrorLocked(attacker, blocker,
				BlockRefusal{Reason: BlockReasonBlockerCapacity, Source: blocker.InstanceID, N: capacity})
		}
		mark(d.Attacker)
	}
	if len(touched) == 0 {
		return entries, nil, nil
	}
	counts := after.blockerCounts()
	for _, atkID := range touched {
		atk := findBattlefieldCard(g, atkID)
		if atk == nil {
			continue
		}
		if err := g.blockCountRefusalLocked(atk, counts[atkID], decls); err != nil {
			return nil, nil, err
		}
	}
	// #1507, CR 509.1b: the whole-combat limit. After the per-attacker
	// bounds so a menace block that is short a creature reports
	// too_few_blockers, the refusal about the declaration itself,
	// rather than a combat-wide one about something else.
	if err := g.blockLimitRefusalLocked(base, after, decls); err != nil {
		return nil, nil, err
	}
	return entries, after, nil
}

// blockDefenderRefusalLocked reports whether `blocker`'s controller is
// the defending player of `attacker`'s attack (CR 802.4a, 509.1a), and
// a not_defending refusal when it is not (#1339).
//
// The defending player is defendingPlayerForAttackerLocked's: the
// player attacked, the controller of the planeswalker attacked, or the
// protector of the battle attacked — and, for an attacker whose
// planeswalker or battle has left the battlefield, the player that was
// when the attack was pointed (CR 506.4c "it may be blocked", CR
// 802.2a; #1364). A creature not attacking at all has no defending
// player, so nobody may block it. blockOptionsLocked reads the same
// function; before #1339 the verb disagreed and took any block.
//
// Separate from BlockPairRefusalLocked on purpose: that function is
// the per-pair CR 509.1b answer about the two CREATURES, and its
// contract says "who controls the blocker" is the declaration's
// business. Its callers other than the declaration (the enumerator,
// the #328 signal) already restrict the blocker set to the seat the
// attack is aimed at.
//
// Source is the attacker: the thing the player needs to look at is
// where it is pointed.
//
// Caller must hold g.mu. Reads only.
func (g *Game) blockDefenderRefusalLocked(attacker, blocker *Card) BlockRefusal {
	if attacker == nil || blocker == nil {
		return BlockRefusal{Reason: BlockReasonNotDefending}
	}
	defender := g.defendingPlayerForAttackerLocked(attacker)
	if defender != uuid.Nil && defender == blocker.Controller {
		return BlockOK
	}
	return BlockRefusal{Reason: BlockReasonNotDefending, Source: attacker.InstanceID}
}

// blockCountRefusalLocked reports why `n` creatures blocking
// `attacker` is an illegal COUNT (CR 509.1b), or nil when it is
// legal. The whole-declaration half of block legality; the per-pair
// half is BlockPairRefusalLocked.
//
// `decls` is the declaration being judged, used only to name a
// blocker from it so the client has a card to highlight. A minimum
// broken by a blocker being re-pointed AWAY names no blocker, which
// is honest: no creature in that declaration is the problem.
//
// Source is the ATTACKER rather than the permanent a bound was read
// from. Menace has no source but the attacker, and blockerBoundsLocked
// combines the bounds of every rule on the battlefield into one pair
// of numbers, so there is no single card to name — the same reason
// BlockRefusal.Source names the affected creature for a restriction
// bit nobody recorded the writer of.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) blockCountRefusalLocked(attacker *Card, n int, decls []BlockDeclaration) *BlockRefusedError {
	// An empty block is always legal: "no blockers" is a valid
	// outcome for any attacker, and a minimum only ever says how many
	// it takes to block at all.
	if attacker == nil || n <= 0 {
		return nil
	}
	min, max := g.blockerBoundsLocked(attacker)
	var r BlockRefusal
	switch {
	case min > 0 && n < min:
		r = BlockRefusal{Reason: BlockReasonTooFewBlockers, Source: attacker.InstanceID, N: min}
	case max > 0 && n > max:
		r = BlockRefusal{Reason: BlockReasonTooManyBlockers, Source: attacker.InstanceID, N: max}
	default:
		return nil
	}
	var blocker *Card
	for _, d := range decls {
		if d.Attacker == attacker.InstanceID {
			blocker = findBattlefieldCard(g, d.Blocker)
			break
		}
	}
	return g.blockRefusedErrorLocked(attacker, blocker, r)
}

// BlockOption is one complete block a seat could declare right now:
// the entries to send as a single DeclareBlockers. A one-entry option
// is an ordinary single block; a longer one is a group that is legal
// only as a group, such as the two creatures a menace attacker takes.
type BlockOption struct {
	Blocks []BlockDeclaration
	// Required marks the block the seat's CR 509.1c requirements are
	// owed (#1597): the witness the checkpoint wants, offered whole so
	// a seat whose pass is refused always has the answer in one move.
	// The enumerator marks it AlwaysLegal.
	Required bool
}

// BlockOptionsLocked is the ONE option generator (ADR 0045 addendum,
// Decision 14): every block `seat` could legally declare right now,
// on top of whatever it has already declared. The legal-move
// enumerator and the #328 auto-pass signal both read it, so neither
// can offer — or hold the window open for — a block the verb would
// refuse.
//
// Two shapes:
//
//   - SINGLES: one entry for each eligible blocker and each attacker
//     it may join alone. That means the pair is legal AND the
//     resulting count is, so a lone block on an unblocked menace
//     attacker is not offered, and neither is a fourth blocker on an
//     attacker capped at three.
//   - MINIMUM GROUPS: for an unblocked attacker whose minimum is
//     m >= 2, combinations of exactly m pair-legal eligible blockers,
//     in battlefield order, capped at perAttackerCap. Blockers beyond
//     the minimum join later as singles.
//
// Every option is run through checkBlockDeclarationLocked before it is
// returned, so a maximum — or any set check added later — can never
// produce an option the engine refuses. That is how a whole-combat
// limit (#1507) stops the generator offering a second blocker once the
// first has used Silent Arbiter's one up: no line here knows about it.
//
// Completeness is not promised (legal.go's contract): perAttackerCap
// bounds the groups per attacker, so on a wide board the best pair may
// not be among them. Soundness is: everything returned is legal.
// perAttackerCap <= 0 means one group per attacker.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) BlockOptionsLocked(seat uuid.UUID, perAttackerCap int) []BlockOption {
	return g.blockOptionsLocked(seat, perAttackerCap, 0)
}

// blockOptionsLocked is BlockOptionsLocked with a bound on how many
// options it builds in total. maxTotal <= 0 builds them all.
//
// The bound exists for the #328 signal, which asks only "is there
// ANY legal block?" and is answered on every snapshot broadcast for
// every seat. Building the whole option set to throw all but one
// away would put a combinatorial walk on that path.
func (g *Game) blockOptionsLocked(seat uuid.UUID, perAttackerCap, maxTotal int) []BlockOption {
	if g.Battlefield == nil || seat == uuid.Nil {
		return nil
	}
	// #1279: a defender whose declaration is complete is offered
	// nothing more. The verb still takes a late block (the sandbox
	// allowance ADR 0045 Decision 38 records), but the enumerator, the
	// bot and the #328 auto-pass signal stop asking — "has this seat
	// still got a block to make" is now "is this seat still declaring".
	if g.blocksDeclared[seat] {
		return nil
	}
	if perAttackerCap <= 0 {
		perAttackerCap = 1
	}
	// Attackers this seat defends. S27: "attacking this seat" is the
	// DEFENDING player of the attack, not a bare id match — an attack
	// on a planeswalker names the walker and is defended by whoever
	// controls it. #1364: and one whose walker has since left is still
	// defended by that player (CR 506.4c).
	var attackers []*Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.AttackingTarget == uuid.Nil {
			continue
		}
		if g.defendingPlayerForAttackerLocked(c) == seat {
			attackers = append(attackers, c)
		}
	}
	if len(attackers) == 0 {
		return nil
	}
	var eligible []*Card
	for i := range g.Battlefield.Cards {
		b := &g.Battlefield.Cards[i]
		// #328: the same per-card eligibility test the wire's
		// block_decision_seats signal uses.
		if BlockerEligible(b, seat) {
			eligible = append(eligible, b)
		}
	}
	if len(eligible) == 0 {
		return nil
	}
	base := g.currentBlockAssignmentLocked()
	var out []BlockOption
	// #1597: the blocks a CR 509.1c requirement is owed come first,
	// as one option. The singles and groups below may not contain them
	// in one piece (a Lure'd menace attacker's three blockers), and a
	// seat that owes a requirement must always be offered — and the
	// #328 signal must always see — a way to obey it.
	if w := g.blockRequirementWitnessLocked(seat); len(w) > 0 {
		out = append(out, BlockOption{Blocks: w, Required: true})
		if maxTotal > 0 && len(out) >= maxTotal {
			return out
		}
	}
	cache := blockReachCache{}
	for _, atk := range attackers {
		// The blockers that may pair with this attacker at all. The
		// count is judged on the sets built from them.
		var pool []*Card
		for _, b := range eligible {
			// #1706: a creature that can block more than one attacker
			// stays eligible while it has room, but is never offered an
			// attacker it already blocks — that entry would change
			// nothing.
			if base.has(b.InstanceID, atk.InstanceID) {
				continue
			}
			if g.CanBlockLocked(atk, b) {
				pool = append(pool, b)
			}
		}
		if len(pool) == 0 {
			continue
		}
		for _, b := range pool {
			decls := []BlockDeclaration{{Blocker: b.InstanceID, Attacker: atk.InstanceID}}
			if _, err := g.checkBlockDeclarationCachedLocked(base, decls, cache); err == nil {
				out = append(out, BlockOption{Blocks: decls})
				if maxTotal > 0 && len(out) >= maxTotal {
					return out
				}
			}
		}
		// Groups only close the gap a minimum opens, and only for an
		// attacker nothing is blocking yet: once it is blocked at all
		// the minimum is met and further blockers are singles.
		min, _ := g.blockerBoundsLocked(atk)
		if min < 2 {
			continue
		}
		if base.blockerCounts()[atk.InstanceID] > 0 {
			continue
		}
		for _, combo := range blockerCombinations(pool, min, perAttackerCap) {
			decls := make([]BlockDeclaration, 0, len(combo))
			for _, b := range combo {
				decls = append(decls, BlockDeclaration{Blocker: b.InstanceID, Attacker: atk.InstanceID})
			}
			if _, err := g.checkBlockDeclarationCachedLocked(base, decls, cache); err == nil {
				out = append(out, BlockOption{Blocks: decls})
				if maxTotal > 0 && len(out) >= maxTotal {
					return out
				}
			}
		}
	}
	return out
}

// blockerCombinations returns up to `limit` combinations of exactly `k`
// cards from `pool`, in the pool's order (which is battlefield order),
// lexicographic by index — so the options a board produces are
// deterministic and a replay reproduces them.
func blockerCombinations(pool []*Card, k, limit int) [][]*Card {
	if k <= 0 || k > len(pool) || limit <= 0 {
		return nil
	}
	var out [][]*Card
	idx := make([]int, k)
	for i := range idx {
		idx[i] = i
	}
	for {
		combo := make([]*Card, k)
		for i, j := range idx {
			combo[i] = pool[j]
		}
		out = append(out, combo)
		if len(out) >= limit {
			return out
		}
		// Advance to the next combination in lexicographic order.
		i := k - 1
		for i >= 0 && idx[i] == i+len(pool)-k {
			i--
		}
		if i < 0 {
			return out
		}
		idx[i]++
		for j := i + 1; j < k; j++ {
			idx[j] = idx[j-1] + 1
		}
	}
}
