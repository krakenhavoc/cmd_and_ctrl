package game

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// block_requirements.go is CR 509.1c on the block declaration — the
// blocking half of the requirement seam #1571 closed for attacks
// (attack_requirements.go). #1597, ADR 0045 amendment of 2026-09-28,
// Decisions 55-59; Decision 15 is the sketch this builds.
//
// # The rule
//
// CR 509.1c: the defending player checks each requirement on the
// creatures they control and on the attacking creatures, and "the
// number of requirements that are being obeyed must be maximized" —
// without disobeying a restriction, and without paying a cost. So,
// exactly as for attacks, a requirement is a property of a complete
// declaration: under Lure three creatures that could block the
// enchanted creature must all block it, while one menace attacker with
// Lure and one potential blocker asks nothing at all, because the lone
// block is illegal (CR 702.111b) and a requirement never beats a
// restriction.
//
// # The five texts
//
//   - "blocks each combat if able" (Watchdog; Grand Melee's second
//     line on every creature) — a requirement on the would-be BLOCKER,
//     obeyed by blocking anything.
//   - "all creatures able to block this creature do so" (Lure, Prized
//     Unicorn) — on the ATTACKER, and it is one requirement on EACH
//     creature able to block it, obeyed by that creature blocking this
//     attacker. So a Lure'd attacker with three potential blockers
//     carries three requirements, and a blocker that goes elsewhere
//     disobeys one.
//   - "must be blocked if able" (Gaea's Protector, Irresistible Prey) —
//     one requirement on the attacker, obeyed by at least one blocker.
//   - "must be blocked by exactly one creature if able" (Nacatl
//     War-Pride) — one requirement, obeyed by exactly one blocker.
//   - "target creature blocks THIS creature this turn if able" (Provoke,
//     Grappling Hook, Turntimber Basilisk; #1684) — a requirement on
//     the would-be BLOCKER that names one attacking OBJECT (instance
//     and epoch, CR 400.7), obeyed only by blocking that attacker.
//   - "it blocks each attacking creature this turn if able" (Blaze of
//     Glory; #1715) — on the would-be BLOCKER, and it is one
//     requirement per ATTACKER it could block, obeyed by blocking that
//     attacker: Lure said from the blocker's side. With a capacity to
//     match (Blaze gives "any number" in the same breath) the most it
//     can obey is every attacker it can legally block; with less, the
//     search fills what capacity it has.
//
// A Lure may be narrowed to some blockers — "All Walls able to block
// this creature do so" (Marble Priest), "All creatures with flying …"
// (Talruum Piper). The narrowing is a registered FILTER KEY on the
// requirement, never a closure (ADR 0041's ratchet: the requirement is
// pure data, compared and copied with the characteristic), and a
// blocker the filter does not match carries no requirement from it.
//
// # How it is judged in an incremental declaration
//
// Exactly the attack side's split (Decision 49), per defending player,
// because each defending player declares on their own (Decision 38):
//
//   - reach(D) is the requirements the staged declaration D obeys plus
//     the best restriction-legal ADDITION of the defending player's
//     unassigned creatures (blockRequirementReachLocked). An empty
//     declaration's reach is CR 509.1c's maximum.
//   - The declaration verb (DeclareBlockers, and DeclareBlocker which is
//     its one-entry form) refuses a declaration that lowers reach.
//   - The CHECKPOINT is where the defender says their declaration is
//     done: their pass in the step, finish_blocks, and AdvanceStep
//     leaving the step. Each is refused while an addition would obey
//     another requirement.
//
// Only a PENDING declaration is judged: once a defender's declaration
// is complete (block_completion.go) a creature that arrives, or a Lure
// that is cast, asks nothing of it.
//
// # The search
//
// Blockers compete for attackers under the count bounds (menace's
// minimum, Hungering Hydra's maximum) and the whole-combat limits
// (Silent Arbiter). Without a minimum every attacker is a capacity and
// the best addition is a min-cost flow, like the attack side's; a
// minimum is not a capacity — an attacker takes 0 blockers or at least
// m — so each attacker with a live minimum, and each with an "exactly
// one" requirement, is given a small set of MODES (closed, exactly one,
// many) and the flow is solved per combination. The witness is then
// COUNTED EXACTLY and re-checked against the ordinary validator, so
// the flow is only ever a search: whatever the refusal or the enumerator
// is handed is a declaration the verb accepts, and the gain it reports
// is the gain it has.

// BlockRequirementKind names one CR 509.1c requirement text. A closed
// vocabulary and an on-disk identity (ModAddBlockRequirement's Text):
// never renamed, never reused.
type BlockRequirementKind string

const (
	// BlockRequirementBlocks is "blocks each combat if able", on the
	// creature that must block.
	BlockRequirementBlocks BlockRequirementKind = "blocks"
	// BlockRequirementLure is "all creatures able to block this
	// creature do so", on the attacker.
	BlockRequirementLure BlockRequirementKind = "lure"
	// BlockRequirementMustBeBlocked is "must be blocked if able", on
	// the attacker.
	BlockRequirementMustBeBlocked BlockRequirementKind = "mustBeBlocked"
	// BlockRequirementExactlyOne is "must be blocked by exactly one
	// creature if able", on the attacker.
	BlockRequirementExactlyOne BlockRequirementKind = "exactlyOne"
	// BlockRequirementBlocksAttacker is "<this creature> blocks <that
	// attacker> this turn if able" (Provoke, Grappling Hook, Turntimber
	// Basilisk; #1684), on the creature that must block. It names ONE
	// attacking object in BlockRequirement.Attacker, and only blocking
	// that object obeys it: blocking anything else obeys nothing, and a
	// new object the attacker's card has become is not it (CR 400.7).
	BlockRequirementBlocksAttacker BlockRequirementKind = "blocksAttacker"
	// BlockRequirementBlocksEach is "it blocks each attacking creature
	// this turn if able" (Blaze of Glory; #1715), on the creature that
	// must block. One requirement per attacker its controller defends
	// against, each obeyed only by blocking THAT attacker — so a
	// creature that can block two of three attackers obeys two and
	// disobeys one, and the declaration must still obey two.
	BlockRequirementBlocksEach BlockRequirementKind = "blocksEach"
)

// KnownBlockRequirementKind reports whether this binary can interpret k.
func KnownBlockRequirementKind(k BlockRequirementKind) bool {
	switch k {
	case BlockRequirementBlocks, BlockRequirementLure, BlockRequirementMustBeBlocked, BlockRequirementExactlyOne,
		BlockRequirementBlocksAttacker, BlockRequirementBlocksEach:
		return true
	}
	return false
}

// blockerSideRequirement reports whether a requirement of kind k sits
// on the creature that must block (as opposed to on the attacker).
func blockerSideRequirement(k BlockRequirementKind) bool {
	return k == BlockRequirementBlocks || k == BlockRequirementBlocksAttacker || k == BlockRequirementBlocksEach
}

// Blocker filter keys (#1684): the registered narrowings a Lure may
// carry in BlockRequirement.Filter. An identity like the kind itself —
// never renamed, never reused — although today only a catalog static
// writes one, so none reaches a snapshot.
const (
	// BlockerFilterWall is "All Walls able to block …" (Marble Priest).
	BlockerFilterWall = "wall"
	// BlockerFilterFlying is "All creatures with flying able to block
	// …" (Talruum Piper).
	BlockerFilterFlying = "flying"
)

// blockerFilters is the registry behind BlockRequirement.Filter. A
// filter reads the would-be blocker as it is now (its effective
// characteristics), because "all Walls able to block" asks about the
// creature at the declaration, not when the Lure began.
var blockerFilters = struct {
	sync.RWMutex
	byKey map[string]func(blocker *Card) bool
}{byKey: map[string]func(blocker *Card) bool{
	BlockerFilterWall:   func(b *Card) bool { return b.HasSubtype("Wall") },
	BlockerFilterFlying: func(b *Card) bool { return HasKeyword(b, "flying") },
}}

// RegisterBlockerFilter adds a blocker filter under `key`. Panics on an
// empty key, a nil predicate or a duplicate — each a card-file bug that
// would otherwise surface as a Lure that silently binds nobody.
func RegisterBlockerFilter(key string, match func(blocker *Card) bool) {
	if key == "" || match == nil {
		panic("game: RegisterBlockerFilter needs a key and a predicate")
	}
	blockerFilters.Lock()
	defer blockerFilters.Unlock()
	if _, dup := blockerFilters.byKey[key]; dup {
		panic(fmt.Sprintf("game: blocker filter %q registered twice", key))
	}
	blockerFilters.byKey[key] = match
}

// KnownBlockerFilter reports whether `key` names a registered filter.
// The empty key is "no filter" and is always known.
func KnownBlockerFilter(key string) bool {
	if key == "" {
		return true
	}
	blockerFilters.RLock()
	defer blockerFilters.RUnlock()
	_, ok := blockerFilters.byKey[key]
	return ok
}

// bindsBlocker reports whether Lure requirement r reaches `blocker`:
// always without a filter, and — with one — only when the registered
// predicate matches. An unregistered key binds nobody, which is the
// weaker-than-printed direction.
func (r BlockRequirement) bindsBlocker(blocker *Card) bool {
	if r.ExceptController != uuid.Nil && (blocker == nil || blocker.Controller == r.ExceptController) {
		return false
	}
	if r.Filter == "" {
		return true
	}
	blockerFilters.RLock()
	match := blockerFilters.byKey[r.Filter]
	blockerFilters.RUnlock()
	return match != nil && blocker != nil && match(blocker)
}

// namesAttacker reports whether BlocksAttacker requirement r names the
// attacking permanent `atk` — the same instance AND the same object
// (CR 400.7).
func (r BlockRequirement) namesAttacker(atk *Card) bool {
	return atk != nil && r.Attacker.ID == atk.InstanceID && r.Attacker.Epoch == atk.ObjectEpoch
}

// blockPairWeight is how many requirements blocker `b` blocking
// attacker `atk` obeys that are PER PAIR — b's "blocks that attacker"
// naming atk, b's "blocks each attacking creature" (#1715), and each
// of atk's Lures that binds b. The search's pair arc cost.
//
// b's "blocks each combat" is NOT here since #1706: blocking anything
// obeys it ONCE, however many attackers a creature that can block
// more takes, so the search charges it on the blocker's first unit
// (blocksEachCombatWeight) rather than on every pair.
func blockPairWeight(b, atk *Card) int {
	n := 0
	for _, r := range blockRequirementsOf(b) {
		if r.Kind == BlockRequirementBlocksAttacker && r.namesAttacker(atk) {
			n++
		}
		if r.Kind == BlockRequirementBlocksEach && atk != nil {
			n++
		}
	}
	for _, r := range blockRequirementsOf(atk) {
		if r.Kind == BlockRequirementLure && r.bindsBlocker(b) {
			n++
		}
	}
	return n
}

// blocksEachCombatWeight is how many "blocks each combat if able"
// requirements `b` carries — each obeyed by b blocking anything.
func blocksEachCombatWeight(b *Card) int {
	return blockRequirementCount(blockRequirementsOf(b), BlockRequirementBlocks)
}

// BlockRequirement is one CR 509.1c requirement on one object. Pure
// data, so the Characteristic that carries it stays copyable.
type BlockRequirement struct {
	Kind BlockRequirementKind
	// Source is the permanent (or the resolved spell or ability's
	// source) whose text imposes it, and SourceName its name as the
	// refusal sentence names it — captured when the requirement is
	// written, because a spell's source is in a graveyard by the time
	// anybody asks.
	Source     uuid.UUID
	SourceName string
	// Attacker is the attacking OBJECT a BlockRequirementBlocksAttacker
	// names (#1684) — instance and epoch, so a creature that left and
	// came back is not it. Zero on every other kind.
	Attacker ObjectRef
	// Filter narrows a BlockRequirementLure to the blockers a registered
	// filter matches (KnownBlockerFilter; "wall", "flying"). Empty on
	// every other kind, and on an ordinary Lure.
	Filter string
	// ExceptController narrows a BlockRequirementLure to blockers NOT
	// controlled by this player (#2050): "all creatures your opponents
	// control able to block it do so" — the writer's controller is
	// exempt. Zero on every other kind and on an ordinary Lure. Read at
	// the declaration, like Filter, so a creature that changed hands
	// since the effect resolved is judged by its controller now.
	ExceptController uuid.UUID
}

// blockRequirementsOf is what the layer pass wrote onto `c`. Read off
// the cache directly: a requirement is only ever written by a static
// or a data record, so a card the layer pass has not reached carries
// none. Caller must hold g.mu with fresh layers.
func blockRequirementsOf(c *Card) []BlockRequirement {
	if c == nil || c.effective == nil {
		return nil
	}
	return c.effective.BlockRequirements
}

// blockRequirementCount is how many of `reqs` are of `kind`.
func blockRequirementCount(reqs []BlockRequirement, kind BlockRequirementKind) int {
	n := 0
	for _, r := range reqs {
		if r.Kind == kind {
			n++
		}
	}
	return n
}

// blockRequirementsJudgedLocked reports whether `defender`'s block
// declaration is past being judged: outside the declare-blockers step,
// not a defending player, or already complete (Decision 38). Nothing
// here asks anything of a judged declaration.
//
// Caller must hold g.mu.
func (g *Game) blockRequirementsJudgedLocked(defender uuid.UUID) bool {
	if g.State != StateActive || g.Turn.Step != StepDeclareBlockers {
		return true
	}
	if g.blocksDeclared[defender] {
		return true
	}
	return !g.isDefendingPlayerLocked(defender)
}

// defendedAttackersLocked lists the attackers `defender` is the
// defending player of, in battlefield order.
//
// Caller must hold g.mu.
func (g *Game) defendedAttackersLocked(defender uuid.UUID) []*Card {
	var out []*Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.AttackingTarget == uuid.Nil {
			continue
		}
		if g.defendingPlayerForAttackerLocked(c) == defender {
			out = append(out, c)
		}
	}
	return out
}

// anyBlockRequirementLocked is the fast path every caller takes first:
// does any creature `defender` controls, or any attacker they defend
// against, carry a requirement at all? False at nearly every table.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) anyBlockRequirementLocked(defender uuid.UUID) bool {
	if defender == uuid.Nil || g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		reqs := blockRequirementsOf(c)
		if len(reqs) == 0 {
			continue
		}
		if c.Controller == defender && c.IsCreature() {
			for _, r := range reqs {
				if blockerSideRequirement(r.Kind) {
					return true
				}
			}
		}
		if c.AttackingTarget != uuid.Nil && g.defendingPlayerForAttackerLocked(c) == defender {
			return true
		}
	}
	return false
}

// blockReqHit is one requirement obeyed by an assignment: the
// requirement, the object carrying it and its index there, and the
// (blocker, attacker) pair that obeys it. For an attacker-side
// requirement Blocker is one of the creatures blocking it, so the
// client has a card to highlight.
type blockReqHit struct {
	req      BlockRequirement
	holder   uuid.UUID
	idx      int
	blocker  uuid.UUID
	attacker uuid.UUID
}

// key identifies the requirement a hit obeys, independent of HOW it is
// obeyed: "blocks each combat" is the same requirement whichever
// attacker the creature blocks, while a Lure requirement is one per
// potential blocker and a "blocks each attacking creature" one per
// attacker (#1715).
func (h blockReqHit) key() blockReqHit {
	k := blockReqHit{holder: h.holder, idx: h.idx}
	switch h.req.Kind {
	case BlockRequirementLure:
		k.blocker = h.blocker
	case BlockRequirementBlocksEach:
		k.attacker = h.attacker
	}
	return k
}

// blockRequirementHitsLocked is every requirement `defender`'s side of
// `assign` (blocker -> attacker) obeys, in battlefield order so the
// refusal a caller sees is stable.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) blockRequirementHitsLocked(defender uuid.UUID, assign blockAssignment) []blockReqHit {
	var out []blockReqHit
	for i := range g.Battlefield.Cards {
		b := &g.Battlefield.Cards[i]
		if b.Controller != defender {
			continue
		}
		atkIDs := assign[b.InstanceID]
		if len(atkIDs) == 0 {
			continue
		}
		// #1706: a creature may block several attackers. Each of its
		// own requirements is still ONE requirement — "blocks each
		// combat" obeyed once however many it blocks, "blocks that
		// attacker" obeyed when the named one is among them — and each
		// Lure is obeyed per (blocker, Lure'd attacker) pair.
		for j, r := range blockRequirementsOf(b) {
			switch r.Kind {
			case BlockRequirementBlocks:
				out = append(out, blockReqHit{req: r, holder: b.InstanceID, idx: j, blocker: b.InstanceID, attacker: atkIDs[0]})
			case BlockRequirementBlocksAttacker:
				for _, atkID := range atkIDs {
					if r.namesAttacker(findBattlefieldCard(g, atkID)) {
						out = append(out, blockReqHit{req: r, holder: b.InstanceID, idx: j, blocker: b.InstanceID, attacker: atkID})
						break
					}
				}
			case BlockRequirementBlocksEach:
				// #1715: one requirement per attacker, so one hit per
				// attacker blocked.
				for _, atkID := range atkIDs {
					if findBattlefieldCard(g, atkID) != nil {
						out = append(out, blockReqHit{req: r, holder: b.InstanceID, idx: j, blocker: b.InstanceID, attacker: atkID})
					}
				}
			}
		}
		for _, atkID := range atkIDs {
			for j, r := range blockRequirementsOf(findBattlefieldCard(g, atkID)) {
				if r.Kind == BlockRequirementLure && r.bindsBlocker(b) {
					out = append(out, blockReqHit{req: r, holder: atkID, idx: j, blocker: b.InstanceID, attacker: atkID})
				}
			}
		}
	}
	for _, atk := range g.defendedAttackersLocked(defender) {
		reqs := blockRequirementsOf(atk)
		if len(reqs) == 0 {
			continue
		}
		n := 0
		var first uuid.UUID
		for i := range g.Battlefield.Cards {
			id := g.Battlefield.Cards[i].InstanceID
			if assign.has(id, atk.InstanceID) {
				if n == 0 {
					first = id
				}
				n++
			}
		}
		for j, r := range reqs {
			switch {
			case r.Kind == BlockRequirementMustBeBlocked && n >= 1,
				r.Kind == BlockRequirementExactlyOne && n == 1:
				out = append(out, blockReqHit{req: r, holder: atk.InstanceID, idx: j, blocker: first, attacker: atk.InstanceID})
			}
		}
	}
	return out
}

// blockReach is one reach computation for one defending player: the
// requirements the staged declaration obeys, the most a legal addition
// obeys on top, and that addition.
type blockReach struct {
	obeyed   int
	addition int
	witness  []BlockDeclaration
}

func (r blockReach) total() int { return r.obeyed + r.addition }

// blockRequirementReachLocked is reach(assign) for `defender`.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) blockRequirementReachLocked(defender uuid.UUID, assign blockAssignment) blockReach {
	out := blockReach{obeyed: len(g.blockRequirementHitsLocked(defender, assign))}
	out.addition, out.witness = g.bestBlockRequirementAdditionLocked(defender, assign, out.obeyed)
	return out
}

// brAttacker and brBlocker are the search's view of the table.
type brAttacker struct {
	card    *Card
	c0      int // blockers already assigned
	min     int
	max     int // 0: unbounded
	mb, one int // must-be-blocked / exactly-one requirement counts
	modes   []brMode
}

type brBlocker struct {
	card    *Card
	limited bool // counted by a whole-combat limit
	// #1706: room is how many more attackers it may block (brInfCap
	// for "any number"); blocking is whether it already blocks one,
	// which means its "blocks each combat" is already obeyed and a
	// whole-combat limit has already counted it.
	room     int
	blocking bool
}

// brMode is how many MORE blockers an attacker may take in one
// combination: none (closed), or at least lo and — when hi > 0 — at
// most hi, with `first` the cost of the first of them. A mode is a
// search device, not a claim: the witness is counted exactly.
type brMode struct {
	closed bool
	lo, hi int
	first  int
}

const (
	brBig     = 1 << 20
	brInfCap  = 1 << 20
	brMaxMode = 256 // combinations searched before the rest are closed
)

// bestBlockRequirementAdditionLocked is the search: the most
// requirements an addition to `assign` of `defender`'s unassigned,
// untapped creatures could obey on top of `obeyed`, and one addition
// that does it. (0, nil) when no legal addition obeys anything more —
// including when the only additions the search finds are ones the
// validator refuses, which makes the sandbox weaker than printed on
// that board, never stronger.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) bestBlockRequirementAdditionLocked(defender uuid.UUID, assign blockAssignment, obeyed int) (int, []BlockDeclaration) {
	var atks []*brAttacker
	counts := assign.blockerCounts()
	for _, c := range g.defendedAttackersLocked(defender) {
		reqs := blockRequirementsOf(c)
		a := &brAttacker{
			card: c,
			c0:   counts[c.InstanceID],
			mb:   blockRequirementCount(reqs, BlockRequirementMustBeBlocked),
			one:  blockRequirementCount(reqs, BlockRequirementExactlyOne),
		}
		a.min, a.max = g.blockerBoundsLocked(c)
		atks = append(atks, a)
	}
	if len(atks) == 0 {
		return 0, nil
	}
	var blks []*brBlocker
	for i := range g.Battlefield.Cards {
		b := &g.Battlefield.Cards[i]
		if b.Controller != defender || !b.IsCreature() || b.Tapped {
			continue
		}
		// An addition never re-points: a creature already blocking
		// takes part only when it can block more (#1706), and only for
		// the room it has left.
		cur := len(assign[b.InstanceID])
		room := brInfCap
		if capacity := BlockCapacity(b); capacity > 0 {
			room = capacity - cur
		}
		if room <= 0 {
			continue
		}
		blks = append(blks, &brBlocker{card: b, room: room, blocking: cur > 0})
	}
	if len(blks) == 0 {
		return 0, nil
	}
	// Pair legality and pair weight, once; -1 marks an illegal pair.
	weight := make([][]int, len(blks))
	anyPair := false
	for i, b := range blks {
		weight[i] = make([]int, len(atks))
		for j, a := range atks {
			weight[i][j] = -1
			if assign.has(b.card.InstanceID, a.card.InstanceID) {
				continue
			}
			if g.BlockPairRefusalLocked(a.card, b.card).Legal() {
				weight[i][j] = blockPairWeight(b.card, a.card)
				anyPair = true
			}
		}
	}
	if !anyPair {
		return 0, nil
	}
	room := g.blockLimitRoomLocked(defender, assign, blks)

	// Modes. An attacker is "special" when a mode choice matters: a
	// live minimum of two or more (0 or at least m), or an "exactly
	// one" requirement (1, or 2 and more).
	var specials []*brAttacker
	for _, a := range atks {
		free := brMode{hi: 0}
		if a.max > 0 {
			free.hi = a.max - a.c0
		}
		if a.c0 == 0 {
			free.first = -a.mb
		}
		need := a.min - a.c0
		if need < 2 && a.one == 0 {
			if a.max > 0 && free.hi <= 0 {
				a.modes = []brMode{{closed: true}}
			} else {
				a.modes = []brMode{free}
			}
			continue
		}
		a.modes = []brMode{{closed: true}}
		if a.one > 0 && a.c0 == 0 && a.min <= 1 {
			a.modes = append(a.modes, brMode{lo: 1, hi: 1, first: -a.mb - a.one})
		}
		many := brMode{lo: need, hi: free.hi}
		if many.lo < 1 {
			many.lo = 1
		}
		if a.one > 0 && many.lo < 2-a.c0 {
			many.lo = 2 - a.c0
		}
		if a.c0 == 0 {
			many.first = -a.mb
		}
		if a.c0 == 1 {
			many.first += a.one
		}
		if a.max == 0 || many.hi >= many.lo {
			a.modes = append(a.modes, many)
		}
		specials = append(specials, a)
	}

	bestGain := 0
	var best []BlockDeclaration
	pick := make([]int, len(specials))
	for tried := 0; tried < brMaxMode; tried++ {
		chosen := map[*brAttacker]brMode{}
		for i, a := range specials {
			chosen[a] = a.modes[pick[i]]
		}
		if w, ok := g.solveBlockRequirementFlow(atks, blks, weight, chosen, room); ok && len(w) > 0 {
			if gain := g.blockAdditionGainLocked(defender, assign, w, obeyed); gain > bestGain {
				bestGain, best = gain, w
			}
		}
		// Next combination (odometer). The combinations past brMaxMode
		// leave the later specials closed — a search that stops early
		// under-counts the maximum, which is the weaker direction.
		k := 0
		for ; k < len(specials); k++ {
			pick[k]++
			if pick[k] < len(specials[k].modes) {
				break
			}
			pick[k] = 0
		}
		if k == len(specials) {
			break
		}
	}
	return bestGain, best
}

// blockAdditionGainLocked is how many more requirements `assign` plus
// `add` obeys than `obeyed` — counted exactly, and 0 when the validator
// would refuse the addition.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) blockAdditionGainLocked(defender uuid.UUID, assign blockAssignment, add []BlockDeclaration, obeyed int) int {
	if _, _, err := g.checkBlockRestrictionsLocked(assign, add); err != nil {
		return 0
	}
	return len(g.blockRequirementHitsLocked(defender, withBlocks(assign, add))) - obeyed
}

// brRoom is the whole-combat limit as the search sees it: how many
// more of `defender`'s creatures the tightest limit lets block, and
// which candidates it counts. Several limits are folded into the
// tightest room over every creature any of them counts — stricter
// than the rules when two limits count different creatures, so the
// witness is always one the validator takes.
type brRoom struct {
	room int // -1: no limit
}

// blockLimitRoomLocked reads every whole-combat block limit against
// `assign` for `defender`'s group and marks the candidates they count.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) blockLimitRoomLocked(defender uuid.UUID, assign blockAssignment, blks []*brBlocker) brRoom {
	out := brRoom{room: -1}
	visit := func(r BlockRule, source *Card) bool {
		if r.Limit == nil {
			return true
		}
		bound := 0
		counted := false
		for _, b := range blks {
			if lim := r.Limit(g, b.card, source); lim > 0 {
				counted = true
				if bound == 0 || lim < bound {
					bound = lim
				}
			}
		}
		if !counted {
			return true
		}
		tally := g.blockLimitTallyLocked(r, source, assign)
		key := uuid.Nil
		if r.LimitPerDefender {
			key = defender
		}
		t := tally[key]
		if t.bound > 0 && t.bound < bound {
			bound = t.bound
		}
		room := bound - t.n
		if room < 0 {
			room = 0
		}
		for _, b := range blks {
			if r.Limit(g, b.card, source) > 0 {
				b.limited = true
			}
		}
		if out.room < 0 || room < out.room {
			out.room = room
		}
		return true
	}
	g.forEachBlockRuleLocked(visit)
	g.forEachScopedBlockRuleLocked(visit)
	return out
}

// solveBlockRequirementFlow runs one min-cost flow for one combination
// of modes and returns its assignment, or ok=false when a mode's
// minimum could not be filled.
//
// The network: source -> [limit room] -> blocker (1, cost -the
// blocker's "blocks" count) -> attacker (1, cost -blockPairWeight: its
// "blocks that attacker" requirements naming this attacker, and the
// attacker's Lures that bind this blocker) -> the attacker's mode arcs
// -> sink. A mode's first lo units cost -brBig each, so the flow fills
// every minimum it can before it spends a blocker on anything else;
// the fill is checked afterwards.
//
// #1706: a blocker that can block more than one attacker gets a second
// source arc for the rest of its room, at cost 0 — its "blocks" count
// is obeyed once, on the first unit. A blocker that already blocks has
// obeyed it, so its first unit costs 0 too, and a whole-combat limit
// has already counted it, so it bypasses the limit node. A limited
// blocker that is not blocking yet keeps a single unit through the
// limit: "counts once, then carries more" is not a flow, and one unit
// is the weaker answer.
func (g *Game) solveBlockRequirementFlow(atks []*brAttacker, blks []*brBlocker, weight [][]int, chosen map[*brAttacker]brMode, room brRoom) ([]BlockDeclaration, bool) {
	nb, na := len(blks), len(atks)
	src, lim := 0, 1
	bNode := func(i int) int { return 2 + i }
	aNode := func(j int) int { return 2 + nb + j }
	sink := 2 + nb + na
	f := newMinCostFlow(sink + 1)
	if room.room >= 0 {
		f.add(src, lim, room.room, 0)
	}
	for i, b := range blks {
		first := 0
		if !b.blocking {
			first = -blocksEachCombatWeight(b.card)
		}
		viaLimit := b.limited && room.room >= 0 && !b.blocking
		if viaLimit {
			f.add(lim, bNode(i), 1, first)
		} else {
			f.add(src, bNode(i), 1, first)
		}
		if b.room > 1 && !viaLimit {
			f.add(src, bNode(i), b.room-1, 0)
		}
	}
	type pairArc struct {
		b, a int
		arc  int
	}
	var pairs []pairArc
	modeOf := func(a *brAttacker) brMode {
		if m, ok := chosen[a]; ok {
			return m
		}
		return a.modes[0]
	}
	for i := range blks {
		for j, a := range atks {
			if weight[i][j] < 0 || modeOf(a).closed {
				continue
			}
			pairs = append(pairs, pairArc{b: i, a: j, arc: f.add(bNode(i), aNode(j), 1, -weight[i][j])})
		}
	}
	type forced struct{ node, arc int }
	var mins []forced
	for j, a := range atks {
		m := modeOf(a)
		if m.closed {
			continue
		}
		left := brInfCap
		if m.hi > 0 {
			left = m.hi
		}
		first := true
		take := func(capacity, cost int) {
			if capacity <= 0 || left <= 0 {
				return
			}
			if capacity > left {
				capacity = left
			}
			left -= capacity
			if first {
				arc := f.add(aNode(j), sink, 1, cost+m.first)
				if cost < 0 {
					mins = append(mins, forced{aNode(j), arc})
				}
				first = false
				capacity--
			}
			if capacity > 0 {
				arc := f.add(aNode(j), sink, capacity, cost)
				if cost < 0 {
					mins = append(mins, forced{aNode(j), arc})
				}
			}
		}
		take(m.lo, -brBig)
		take(brInfCap, 0)
	}
	f.run(src, sink)
	for _, fm := range mins {
		if f.g[fm.node][fm.arc].cap != 0 {
			return nil, false
		}
	}
	var out []BlockDeclaration
	for _, p := range pairs {
		if f.g[bNode(p.b)][p.arc].cap == 0 {
			out = append(out, BlockDeclaration{Blocker: blks[p.b].card.InstanceID, Attacker: atks[p.a].card.InstanceID})
		}
	}
	return out, true
}

// minCostFlow is a small successive-shortest-path min-cost flow with
// Bellman-Ford, for networks of a few dozen nodes. It pushes one unit
// at a time along the cheapest path while that path still gains
// (costs < 0), which is the maximum-weight flow of any size. The
// starting network is acyclic, so the residual graph never holds a
// negative cycle.
type minCostFlow struct {
	g [][]mcfArc
}

type mcfArc struct{ to, rev, cap, cost int }

func newMinCostFlow(n int) *minCostFlow { return &minCostFlow{g: make([][]mcfArc, n)} }

// add adds an arc and returns its index in g[u].
func (f *minCostFlow) add(u, v, capacity, cost int) int {
	f.g[u] = append(f.g[u], mcfArc{to: v, rev: len(f.g[v]), cap: capacity, cost: cost})
	f.g[v] = append(f.g[v], mcfArc{to: u, rev: len(f.g[u]) - 1, cap: 0, cost: -cost})
	return len(f.g[u]) - 1
}

func (f *minCostFlow) run(src, sink int) {
	n := len(f.g)
	const inf = 1 << 60
	dist := make([]int, n)
	prevNode := make([]int, n)
	prevArc := make([]int, n)
	for {
		for i := range dist {
			dist[i] = inf
			prevNode[i] = -1
		}
		dist[src] = 0
		for iter := 0; iter < n; iter++ {
			changed := false
			for u := 0; u < n; u++ {
				if dist[u] == inf {
					continue
				}
				for k, a := range f.g[u] {
					if a.cap > 0 && dist[u]+a.cost < dist[a.to] {
						dist[a.to] = dist[u] + a.cost
						prevNode[a.to] = u
						prevArc[a.to] = k
						changed = true
					}
				}
			}
			if !changed {
				break
			}
		}
		if dist[sink] >= 0 {
			return
		}
		for v := sink; v != src; v = prevNode[v] {
			u := prevNode[v]
			a := &f.g[u][prevArc[v]]
			a.cap--
			f.g[v][a.rev].cap++
		}
	}
}

// blockReachCache memoises reach(base) per defending player for one
// fixed base — the option generator judges dozens of declarations
// against the same starting point.
type blockReachCache map[uuid.UUID]blockReach

// blockRequirementRefusalLocked is the verb-side check (CR 509.1c):
// nil when staging `after` in place of `base` keeps every requirement
// that could still be obeyed obeyable for each pending defending
// player the declaration touches, and the refusal naming one it makes
// unobeyable otherwise.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) blockRequirementRefusalLocked(base, after blockAssignment, decls []BlockDeclaration, cache blockReachCache) *BlockRefusedError {
	var seen []uuid.UUID
	for _, d := range decls {
		b := findBattlefieldCard(g, d.Blocker)
		if b == nil {
			continue
		}
		def := b.Controller
		dup := false
		for _, s := range seen {
			if s == def {
				dup = true
				break
			}
		}
		if dup {
			continue
		}
		seen = append(seen, def)
		if g.blockRequirementsJudgedLocked(def) || !g.anyBlockRequirementLocked(def) {
			continue
		}
		rb, ok := cache[def]
		if !ok {
			rb = g.blockRequirementReachLocked(def, base)
			if cache != nil {
				cache[def] = rb
			}
		}
		ra := g.blockRequirementReachLocked(def, after)
		if ra.total() >= rb.total() {
			continue
		}
		return g.explainBlockReachDropLocked(def, withBlocks(base, rb.witness), withBlocks(after, ra.witness))
	}
	return nil
}

// explainBlockReachDropLocked names a requirement the plan `could`
// obeys and the plan `now` does not — the one the refused declaration
// (or pass) gives up.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) explainBlockReachDropLocked(defender uuid.UUID, could, now blockAssignment) *BlockRefusedError {
	kept := map[blockReqHit]bool{}
	for _, h := range g.blockRequirementHitsLocked(defender, now) {
		kept[h.key()] = true
	}
	for _, h := range g.blockRequirementHitsLocked(defender, could) {
		if !kept[h.key()] {
			return g.blockRequirementErrorLocked(h)
		}
	}
	return &BlockRefusedError{BlockRefusal: BlockRefusal{Reason: BlockReasonRequirement}}
}

// blockRequirementErrorLocked is the refusal for one requirement hit.
//
// Caller must hold g.mu. Reads only.
func (g *Game) blockRequirementErrorLocked(h blockReqHit) *BlockRefusedError {
	e := g.blockRefusedErrorLocked(findBattlefieldCard(g, h.attacker), findBattlefieldCard(g, h.blocker),
		BlockRefusal{Reason: BlockReasonRequirement, Source: h.req.Source})
	e.Requirement = h.req
	e.SourceName = h.req.SourceName
	return e
}

// blockRequirementsUnmetLocked is the checkpoint's question for one
// defending player: the refusal for their declaration as it stands,
// or nil when no legal addition would obey another requirement (or
// the declaration is already judged).
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) blockRequirementsUnmetLocked(defender uuid.UUID) *BlockRefusedError {
	w := g.blockRequirementWitnessLocked(defender)
	if len(w) == 0 {
		return nil
	}
	base := g.currentBlockAssignmentLocked()
	return g.explainBlockReachDropLocked(defender, withBlocks(base, w), base)
}

// blockRequirementWitnessLocked is the addition the checkpoint wants
// from `defender`: the blocks that would obey the most requirements
// still obeyable, or nil when none is owed.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) blockRequirementWitnessLocked(defender uuid.UUID) []BlockDeclaration {
	if g.blockRequirementsJudgedLocked(defender) || !g.anyBlockRequirementLocked(defender) {
		return nil
	}
	r := g.blockRequirementReachLocked(defender, g.currentBlockAssignmentLocked())
	if r.addition <= 0 {
		return nil
	}
	return r.witness
}

// BlockRequirementsUnmetForEffect is blockRequirementsUnmetLocked for
// the legal-move enumerator, which withholds the defending player's
// pass while it is non-nil (#1597).
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) BlockRequirementsUnmetForEffect(defender uuid.UUID) *BlockRefusedError {
	return g.blockRequirementsUnmetLocked(defender)
}

// MustBlockForEffect lists the creatures a pending defending player
// owes a block with right now — the witness the checkpoint wants,
// blocker -> attacker (the first one, for a creature the witness has
// blocking several — #1706), across every defending player. It is the view's
// `must_block` stamp and the same blocks the enumerator offers as the
// declaration's required answer, so the two agree by construction.
// Nil whenever every pass in the step would be accepted.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) MustBlockForEffect() map[uuid.UUID]uuid.UUID {
	if g.Turn.Step != StepDeclareBlockers {
		return nil
	}
	var out map[uuid.UUID]uuid.UUID
	for _, s := range g.Seats {
		if s == nil || s.Eliminated {
			continue
		}
		for _, d := range g.blockRequirementWitnessLocked(s.ID) {
			if out == nil {
				out = map[uuid.UUID]uuid.UUID{}
			}
			if _, seen := out[d.Blocker]; !seen {
				out[d.Blocker] = d.Attacker
			}
		}
	}
	return out
}

// blockCheckpointLocked is the checkpoint itself for `defender`, run
// where their declaration would complete by their own say-so — their
// pass in the step, finish_blocks, and AdvanceStep leaving the step.
//
// Caller must hold g.mu in write mode.
func (g *Game) blockCheckpointLocked(defender uuid.UUID) error {
	if g.blockRequirementsJudgedLocked(defender) {
		return nil
	}
	g.RecomputeLayersIfStaleLocked()
	if e := g.blockRequirementsUnmetLocked(defender); e != nil {
		return e
	}
	return nil
}
