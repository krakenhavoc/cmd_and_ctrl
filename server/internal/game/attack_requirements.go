package game

import (
	"errors"

	"github.com/google/uuid"
)

// attack_requirements.go is CR 508.1d on the attack declaration:
// "attacks each combat if able", "creatures your opponents control
// attack this turn if able" (Bident of Thassa), and goad's two
// requirements — "attacks each combat if able and attacks a player
// other than the goading player if able" (CR 701.15b). #1571, ADR 0045
// amendment of 2026-09-24, Decisions 48-52.
//
// # The rule, and why it is counted over a whole declaration
//
// CR 508.1d: the active player checks every requirement on the
// creatures they could attack with, and "the number of requirements
// that are being obeyed must be maximized" — the maximum possible
// "without disobeying any restrictions". A requirement that would need
// a COST paid is not counted toward that maximum (the same rule, "if
// a creature can't attack unless a cost is paid, that player isn't
// required to pay that cost").
//
// That is a property of a complete declaration, never of one creature:
// under Silent Arbiter two creatures that must attack cannot both, and
// either alone is legal. So nothing here refuses "this creature" — it
// refuses a DECLARATION that obeys fewer requirements than it could.
//
// # How it is judged in an engine whose declaration is incremental
//
// The declaration verbs stage one creature (DeclareAttacker) or a set
// (DeclareAttackers) at a time, and the active player says "done" by
// passing priority (attackers.go). So the rule is split along that
// line, exactly as ADR 0045 Decision 15 sketched for blocking
// requirements:
//
//   - reach(D) is the most requirements any completion of the staged
//     declaration D could obey: the requirements D already obeys plus
//     the best FREE, restriction-legal addition of the creatures that
//     are not attacking yet (attackRequirementReachLocked). An empty
//     declaration's reach is CR 508.1d's maximum.
//   - A declaration VERB is refused when it lowers reach — when, after
//     it, some requirement that could still have been obeyed no longer
//     can. Declaring a goaded creature at its goader while another
//     opponent is open, or a creature with no requirement into Silent
//     Arbiter's only slot while one that must attack waits. Every verb
//     that is accepted keeps reach where it was, so reach never falls
//     below the maximum.
//   - The CHECKPOINT is the active player's pass in the step (and
//     AdvanceStep leaving it): refused while a free addition would
//     still obey another requirement. Because reach never fell, a
//     declaration that passes the checkpoint obeys the maximum.
//
// Judging reach rather than recomputing the maximum from scratch at
// the pass keeps the sandbox from wedging: a board that changes under
// a half-made declaration (a goad cast mid-step, an attack limit that
// arrives) can only ever ask for more ADDITIONS, never demand that the
// player unwind something already staged, and the legal-move
// enumerator always has an addition to offer when the pass is withheld.
//
// # What a requirement is
//
// Card-side requirements ride Characteristic.AttackRequirements,
// written by ordinary layer statics (Zurgo's own, Grand Melee's on
// every creature, Bident's floating one on "creatures your opponents
// control"), so the layer pass decides WHO is affected and a duration
// decides WHEN it ends — no second vocabulary. Goad's pair is read off
// Card.GoadedBy here, because goad is a per-object marker with its own
// lifetime (CR 701.15a), not a continuous effect a static describes.
//
// # The search
//
// Without an attack limit every creature is independent and the
// maximum is a sum of per-creature bests. With one (Silent Arbiter,
// Crawlspace, The Eternal Wanderer) creatures compete for slots, and
// the most requirements obeyable is a maximum-weight assignment of
// creatures to attack targets under per-target and combat-wide
// capacities. Every limit counts either every attack or the attacks on
// one target (attack_limits.go), so the capacities are a two-level
// family and a small min-cost flow solves it exactly
// (maxRequirementAssignment). Only creatures that carry a requirement
// take part: one with none adds nothing, and leaving it home is always
// at least as good.

// AttackRequirement is one CR 508.1d requirement on one creature.
//
// Two shapes, told apart by OtherThan:
//
//   - OtherThan == uuid.Nil: "attacks each combat if able" — obeyed by
//     attacking anything.
//   - OtherThan set: "attacks a player other than OtherThan if able"
//     (goad, Kardur) — obeyed only by attacking a PLAYER who is not
//     OtherThan. An attack on a planeswalker or a battle obeys the
//     plain requirement and not this one, which is what makes a
//     goaded creature attack a player rather than a walker when it
//     can (CR 701.15b says "player").
//
// Pure data, so the Characteristic that carries it stays copyable.
type AttackRequirement struct {
	// Source is the permanent (or the resolved spell or ability's
	// source) whose text imposes the requirement; uuid.Nil for goad.
	Source uuid.UUID
	// SourceName is that source's name as the refusal sentence names
	// it. Captured when the requirement is written, because a floating
	// effect's source is usually in a graveyard by the time anybody
	// asks.
	SourceName string
	// GoadedBy is the goading player, for goad's two requirements
	// (CR 701.15b); uuid.Nil for every other source.
	GoadedBy uuid.UUID
	// OtherThan, when set, makes this "attacks a player other than
	// OtherThan if able".
	OtherThan uuid.UUID
}

// obeyedBy reports whether an attack at `target` obeys r. uuid.Nil is
// "not attacking", which obeys nothing.
//
// Caller must hold g.mu.
func (r AttackRequirement) obeyedBy(g *Game, target uuid.UUID) bool {
	if target == uuid.Nil {
		return false
	}
	if r.OtherThan == uuid.Nil {
		return true
	}
	return target != r.OtherThan && g.classifyAttackTargetLocked(target) == AttackTargetPlayer
}

// attackRequirementsOfLocked is every requirement on `c` right now:
// what the layer pass wrote onto its effective characteristic, plus
// goad's pair (CR 701.15b).
//
// Goad adds TWO requirements, not one, and the difference is
// observable under CR 508.1d's counting: attacking the goader obeys
// one of them, attacking anybody else obeys both, so a goaded creature
// that can reach another player must.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) attackRequirementsOfLocked(c *Card) []AttackRequirement {
	if c == nil {
		return nil
	}
	var out []AttackRequirement
	if eff := c.Effective(); len(eff.AttackRequirements) > 0 {
		out = append(out, eff.AttackRequirements...)
	}
	if c.GoadedBy != uuid.Nil {
		out = append(out,
			AttackRequirement{GoadedBy: c.GoadedBy},
			AttackRequirement{GoadedBy: c.GoadedBy, OtherThan: c.GoadedBy},
		)
	}
	return out
}

// requirementWeightLocked is how many of `reqs` an attack at `target`
// obeys. Caller must hold g.mu.
func (g *Game) requirementWeightLocked(reqs []AttackRequirement, target uuid.UUID) int {
	n := 0
	for _, r := range reqs {
		if r.obeyedBy(g, target) {
			n++
		}
	}
	return n
}

// attackRequirementsJudgedLocked reports whether this combat's attack
// declaration has already passed its checkpoint, after which nothing
// here judges anything: the rules' declaration is over, and a creature
// that arrives, or is goaded, afterwards was never asked to attack
// (CR 508.1d applies at the declaration). Also true outside the
// declare-attackers step.
//
// Caller must hold g.mu.
func (g *Game) attackRequirementsJudgedLocked() bool {
	return g.Turn.Step != StepDeclareAttackers || g.attacksDeclared
}

// reqCandidate is one of the declaring player's creatures that carries
// a requirement: the targets it could attack FOR FREE, and how many of
// its requirements each would obey. Pairs that obey nothing are left
// out — they never help the search.
type reqCandidate struct {
	id    uuid.UUID
	name  string
	reqs  []AttackRequirement
	pairs []reqPair
}

type reqPair struct {
	target uuid.UUID
	weight int
}

// attackRequirementCandidatesLocked lists the declaring player's
// creatures that carry at least one requirement and are NOT in
// `assign`, with every free, legal target that obeys something. Only
// AttackerEligible creatures are candidates — a tapped, summoning-sick
// or "can't attack" creature cannot be asked to attack, which is
// CR 508.1d's "without disobeying any restrictions".
//
// "Free" is CR 508.1d's cost clause: a pair whose CR 508.1a tax is not
// zero is not counted, so a requirement never forces a payment.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) attackRequirementCandidatesLocked(ap uuid.UUID, assign map[uuid.UUID]uuid.UUID) []reqCandidate {
	if ap == uuid.Nil || g.Battlefield == nil {
		return nil
	}
	var targets []AttackTargetRef
	var out []reqCandidate
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if _, attacking := assign[c.InstanceID]; attacking {
			continue
		}
		if !AttackerEligible(c, ap) {
			continue
		}
		reqs := g.attackRequirementsOfLocked(c)
		if len(reqs) == 0 {
			continue
		}
		if targets == nil {
			targets = g.AttackTargetsForEffect(ap)
		}
		cand := reqCandidate{id: c.InstanceID, name: c.Effective().Name, reqs: reqs}
		for _, t := range targets {
			w := g.requirementWeightLocked(reqs, t.ID)
			if w == 0 {
				continue
			}
			if !g.priceAttackDeclarationLocked([]AttackDeclaration{{Attacker: c.InstanceID, Target: t.ID}}).IsFree() {
				continue
			}
			cand.pairs = append(cand.pairs, reqPair{target: t.ID, weight: w})
		}
		if len(cand.pairs) > 0 {
			out = append(out, cand)
		}
	}
	return out
}

// attackCapacity is how many MORE creatures the attack limits on the
// battlefield let attack, given the attacks in some assignment: a
// combat-wide room (-1 for unlimited) and a room per target.
type attackCapacity struct {
	global    int
	perTarget map[uuid.UUID]int
}

// attackCapacityLocked reads every active attack limit against
// `assign`. A limit already over its Max (it arrived late, Decision
// 44) leaves room 0.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) attackCapacityLocked(assign map[uuid.UUID]uuid.UUID) attackCapacity {
	cp := attackCapacity{global: -1}
	for _, l := range g.activeAttackLimitsLocked() {
		room := l.limit.Max - l.countIn(assign)
		if room < 0 {
			room = 0
		}
		var key uuid.UUID
		switch l.limit.Scope {
		case AttackLimitEachCombat:
			if cp.global < 0 || room < cp.global {
				cp.global = room
			}
			continue
		case AttackLimitAttackingYou:
			if l.source == nil {
				continue
			}
			key = l.source.Controller
		case AttackLimitAttackingThis:
			if l.source == nil {
				continue
			}
			key = l.source.InstanceID
		default:
			continue
		}
		if cp.perTarget == nil {
			cp.perTarget = map[uuid.UUID]int{}
		}
		if cur, ok := cp.perTarget[key]; !ok || room < cur {
			cp.perTarget[key] = room
		}
	}
	return cp
}

// maxRequirementAssignment is the most requirements the candidates can
// obey together under `cp`, and one assignment that does it (candidate
// id -> target; a candidate left home is absent).
//
// A min-cost flow, source -> creature (1) -> target (1, cost -weight)
// -> target's room -> combat-wide room -> sink, augmented one creature
// at a time along the cheapest path while that path still gains
// something. Every arc out of the source carries 1, so each
// augmentation seats one creature and the loop runs at most once per
// candidate; with no negative cycle in the starting network the
// residual graph never grows one, which is what keeps Bellman-Ford
// exact here.
func maxRequirementAssignment(cands []reqCandidate, cp attackCapacity) (int, map[uuid.UUID]uuid.UUID) {
	if len(cands) == 0 {
		return 0, nil
	}
	// Target nodes, in first-seen order so the witness is deterministic.
	var targets []uuid.UUID
	tIndex := map[uuid.UUID]int{}
	for _, c := range cands {
		for _, p := range c.pairs {
			if _, ok := tIndex[p.target]; !ok {
				tIndex[p.target] = len(targets)
				targets = append(targets, p.target)
			}
		}
	}
	n, m := len(cands), len(targets)
	src, gnode, sink := 0, 1+n+m, 2+n+m
	nodes := sink + 1
	const inf = 1 << 30

	type arc struct{ to, rev, cap, cost int }
	graph := make([][]arc, nodes)
	addArc := func(u, v, capacity, cost int) {
		graph[u] = append(graph[u], arc{to: v, rev: len(graph[v]), cap: capacity, cost: cost})
		graph[v] = append(graph[v], arc{to: u, rev: len(graph[u]) - 1, cap: 0, cost: -cost})
	}
	for i, c := range cands {
		addArc(src, 1+i, 1, 0)
		for _, p := range c.pairs {
			addArc(1+i, 1+n+tIndex[p.target], 1, -p.weight)
		}
	}
	for j, t := range targets {
		room := inf
		if r, ok := cp.perTarget[t]; ok {
			room = r
		}
		addArc(1+n+j, gnode, room, 0)
	}
	global := inf
	if cp.global >= 0 {
		global = cp.global
	}
	addArc(gnode, sink, global, 0)

	total := 0
	for {
		dist := make([]int, nodes)
		prevNode := make([]int, nodes)
		prevArc := make([]int, nodes)
		for i := range dist {
			dist[i] = inf
			prevNode[i] = -1
		}
		dist[src] = 0
		for iter := 0; iter < nodes; iter++ {
			changed := false
			for u := 0; u < nodes; u++ {
				if dist[u] == inf {
					continue
				}
				for k, a := range graph[u] {
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
			break
		}
		for v := sink; v != src; v = prevNode[v] {
			u := prevNode[v]
			a := &graph[u][prevArc[v]]
			a.cap--
			graph[v][a.rev].cap++
		}
		total -= dist[sink]
	}
	witness := map[uuid.UUID]uuid.UUID{}
	for i, c := range cands {
		for _, a := range graph[1+i] {
			if a.to > n && a.to <= n+m && a.cap == 0 && a.cost < 0 {
				witness[c.id] = targets[a.to-1-n]
			}
		}
	}
	return total, witness
}

// attackRequirementReach is the result of one reach computation: the
// requirements the staged declaration obeys, the best free addition,
// and the witness addition.
type attackRequirementReach struct {
	obeyed   int
	addition int
	witness  map[uuid.UUID]uuid.UUID
	cands    []reqCandidate
}

func (r attackRequirementReach) total() int { return r.obeyed + r.addition }

// attackRequirementReachLocked is reach(assign): the requirements the
// declaring player's attacking creatures obey in `assign`, plus the
// most a free, restriction-legal addition of their other creatures
// could obey on top — with every attack in `assign`, whoever controls
// it, counted against the attack limits.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) attackRequirementReachLocked(ap uuid.UUID, assign map[uuid.UUID]uuid.UUID) attackRequirementReach {
	var out attackRequirementReach
	for id, target := range assign {
		c := findBattlefieldCard(g, id)
		if c == nil || c.Controller != ap {
			continue
		}
		out.obeyed += g.requirementWeightLocked(g.attackRequirementsOfLocked(c), target)
	}
	out.cands = g.attackRequirementCandidatesLocked(ap, assign)
	out.addition, out.witness = maxRequirementAssignment(out.cands, g.attackCapacityLocked(assign))
	return out
}

// anyAttackRequirementLocked is the fast path every caller takes
// first: does any creature the declaring player controls carry a
// requirement at all? False at nearly every table, and then nothing
// below builds an assignment.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) anyAttackRequirementLocked(ap uuid.UUID) bool {
	if ap == uuid.Nil || g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != ap || !c.IsCreature() {
			continue
		}
		if c.GoadedBy != uuid.Nil || len(c.Effective().AttackRequirements) > 0 {
			return true
		}
	}
	return false
}

// ErrAttackRequirement is returned when an attack declaration, or the
// active player's pass that ends it, would obey fewer CR 508.1d
// requirements than it could. The error actually returned is an
// *AttackRequirementError naming the creature and the requirement.
// Nothing was declared, tapped or paid.
var ErrAttackRequirement = errors.New("game: the attack declaration does not obey every requirement it could")

// AttackRequirementError names the requirement a refused declaration
// (or pass) leaves unobeyed although it could be obeyed. It wraps
// ErrAttackRequirement.
type AttackRequirementError struct {
	// Attacker is the creature carrying the requirement, and
	// AttackerName its name.
	Attacker     uuid.UUID
	AttackerName string
	// Requirement is the requirement itself.
	Requirement AttackRequirement
	// GoaderName / OtherThanName are the display names of the players
	// Requirement names, for the sentence.
	GoaderName    string
	OtherThanName string
}

// Error is the debug form.
func (e *AttackRequirementError) Error() string {
	return "game: " + e.Sentence(uuid.Nil)
}

// Unwrap makes errors.Is(err, ErrAttackRequirement) hold.
func (e *AttackRequirementError) Unwrap() error { return ErrAttackRequirement }

// Sentence is the player-facing explanation addressed to `viewer`,
// built server-side so the client never re-derives a requirement
// (ADR 0045 §6):
//
//	"Zurgo Helmsmasher must attack this combat if able."
//	"Grizzly Bears must attack this combat if able (Grand Melee)."
//	"Grizzly Bears is goaded by you and must attack a player other than you if able."
func (e *AttackRequirementError) Sentence(viewer uuid.UUID) string {
	who := nameOr(e.AttackerName, "That creature")
	r := e.Requirement
	player := func(id uuid.UUID, name string) string {
		if id != uuid.Nil && id == viewer {
			return "you"
		}
		return nameOr(name, "that player")
	}
	if r.GoadedBy != uuid.Nil {
		s := who + " is goaded by " + player(r.GoadedBy, e.GoaderName)
		if r.OtherThan != uuid.Nil {
			return s + " and must attack a player other than " + player(r.OtherThan, e.OtherThanName) + " if able."
		}
		return s + " and must attack this combat if able."
	}
	var s string
	if r.OtherThan != uuid.Nil {
		s = who + " must attack a player other than " + player(r.OtherThan, e.OtherThanName) + " if able"
	} else {
		s = who + " must attack this combat if able"
	}
	if r.SourceName != "" && r.SourceName != e.AttackerName {
		s += " (" + r.SourceName + ")"
	}
	return s + "."
}

// attackRequirementErrorLocked builds the refusal for creature `id`,
// naming the first of its requirements that `lost` (the target it now
// has, or uuid.Nil) fails and `could` (the target it could have had)
// obeys.
//
// Caller must hold g.mu.
func (g *Game) attackRequirementErrorLocked(id uuid.UUID, reqs []AttackRequirement, lost, could uuid.UUID) *AttackRequirementError {
	e := &AttackRequirementError{Attacker: id}
	if c := findBattlefieldCard(g, id); c != nil {
		e.AttackerName = c.Effective().Name
	}
	for _, r := range reqs {
		if r.obeyedBy(g, could) && !r.obeyedBy(g, lost) {
			e.Requirement = r
			break
		}
	}
	if p := g.playerByIDLocked(e.Requirement.GoadedBy); p != nil {
		e.GoaderName = p.Name
	}
	if p := g.playerByIDLocked(e.Requirement.OtherThan); p != nil {
		e.OtherThanName = p.Name
	}
	return e
}

// explainReachDropLocked names the requirement a reach drop from
// `before` to `after` costs: a creature the before-witness (or the
// before-declaration) had obeying more than the after-state lets it.
//
// Caller must hold g.mu.
func (g *Game) explainReachDropLocked(ap uuid.UUID, beforeAssign, afterAssign map[uuid.UUID]uuid.UUID, before, after attackRequirementReach) *AttackRequirementError {
	planned := func(assign map[uuid.UUID]uuid.UUID, r attackRequirementReach, id uuid.UUID) uuid.UUID {
		if t, ok := assign[id]; ok {
			return t
		}
		return r.witness[id]
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != ap {
			continue
		}
		reqs := g.attackRequirementsOfLocked(c)
		if len(reqs) == 0 {
			continue
		}
		could := planned(beforeAssign, before, c.InstanceID)
		now := planned(afterAssign, after, c.InstanceID)
		if g.requirementWeightLocked(reqs, could) > g.requirementWeightLocked(reqs, now) {
			return g.attackRequirementErrorLocked(c.InstanceID, reqs, now, could)
		}
	}
	// The drop is spread across creatures in a way no single one
	// shows (two creatures trading one limit slot); name the first
	// creature the before-plan used.
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if t, ok := before.witness[c.InstanceID]; ok {
			return g.attackRequirementErrorLocked(c.InstanceID, g.attackRequirementsOfLocked(c), uuid.Nil, t)
		}
	}
	return &AttackRequirementError{}
}

// attackRequirementRefusalLocked is the verb-side check (CR 508.1d):
// nil when staging `decls` on top of the current declaration keeps
// every requirement that could still be obeyed obeyable, and the
// refusal naming one that it makes unobeyable otherwise.
//
// Only judged for the active player's creatures, inside
// declare_attackers, before the declaration's checkpoint. An entry
// naming a creature that is already attacking RE-POINTS it, as the
// verbs do.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) attackRequirementRefusalLocked(decls []AttackDeclaration) *AttackRequirementError {
	if len(decls) == 0 || g.attackRequirementsJudgedLocked() {
		return nil
	}
	ap := g.activePlayerIDLocked()
	if !g.anyAttackRequirementLocked(ap) {
		return nil
	}
	before := g.currentAttackAssignmentLocked()
	after := make(map[uuid.UUID]uuid.UUID, len(before)+len(decls))
	for a, t := range before {
		after[a] = t
	}
	for _, d := range decls {
		after[d.Attacker] = d.Target
	}
	rb := g.attackRequirementReachLocked(ap, before)
	ra := g.attackRequirementReachLocked(ap, after)
	if ra.total() >= rb.total() {
		return nil
	}
	return g.explainReachDropLocked(ap, before, after, rb, ra)
}

// AttackRequirementRefusalForEffect is attackRequirementRefusalLocked
// for the legal-move enumerator: the moves it offers are exactly the
// ones the verbs accept (#544).
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) AttackRequirementRefusalForEffect(decls []AttackDeclaration) *AttackRequirementError {
	return g.attackRequirementRefusalLocked(decls)
}

// attackRequirementsUnmetLocked is the checkpoint's question: the
// refusal for the active player's declaration as it stands, or nil
// when no free addition would obey another requirement (or the
// declaration has already been judged).
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) attackRequirementsUnmetLocked() *AttackRequirementError {
	if g.attackRequirementsJudgedLocked() {
		return nil
	}
	ap := g.activePlayerIDLocked()
	if !g.anyAttackRequirementLocked(ap) {
		return nil
	}
	r := g.attackRequirementReachLocked(ap, g.currentAttackAssignmentLocked())
	if r.addition == 0 {
		return nil
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		t, ok := r.witness[c.InstanceID]
		if !ok {
			continue
		}
		return g.attackRequirementErrorLocked(c.InstanceID, g.attackRequirementsOfLocked(c), uuid.Nil, t)
	}
	return &AttackRequirementError{}
}

// AttackRequirementsUnmetForEffect is attackRequirementsUnmetLocked
// for the enumerator (which withholds the active player's pass while
// it is non-nil) and the view.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) AttackRequirementsUnmetForEffect() *AttackRequirementError {
	return g.attackRequirementsUnmetLocked()
}

// MustAttackForEffect lists the creatures the active player owes an
// attack with right now: while the checkpoint would refuse the pass,
// every eligible, not-yet-attacking creature that could be declared
// at a free target obeying one of its requirements without lowering
// reach. Under Silent Arbiter with two creatures that must attack,
// both are listed until one of them attacks — either is a legal
// answer. Empty whenever the pass would be accepted.
//
// It is the view's `must_attack` stamp and the enumerator's set of
// "required" attack moves, so the two agree by construction.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) MustAttackForEffect() map[uuid.UUID][]uuid.UUID {
	if g.attackRequirementsUnmetLocked() == nil {
		return nil
	}
	ap := g.activePlayerIDLocked()
	before := g.currentAttackAssignmentLocked()
	base := g.attackRequirementReachLocked(ap, before)
	out := map[uuid.UUID][]uuid.UUID{}
	for _, c := range base.cands {
		for _, p := range c.pairs {
			after := make(map[uuid.UUID]uuid.UUID, len(before)+1)
			for a, t := range before {
				after[a] = t
			}
			after[c.id] = p.target
			if g.attackLimitRefusalLocked([]AttackDeclaration{{Attacker: c.id, Target: p.target}}) != nil {
				continue
			}
			if g.attackRequirementReachLocked(ap, after).total() < base.total() {
				continue
			}
			out[c.id] = append(out[c.id], p.target)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// markAttackDeclarationJudgedLocked records that this combat's attack
// declaration has passed its checkpoint. Cleared with the rest of
// combat (clearAttackAnnouncementsLocked).
//
// Caller must hold g.mu in write mode.
func (g *Game) markAttackDeclarationJudgedLocked() {
	if g.Turn.Step == StepDeclareAttackers {
		g.attacksDeclared = true
	}
}

// attackCheckpointLocked is the checkpoint itself, run where the
// active player says the declaration is done — their pass in the
// step, and AdvanceStep leaving it. Refuses with the unmet
// requirement; otherwise records the declaration as judged.
//
// Caller must hold g.mu in write mode, with the cursor in
// declare_attackers.
func (g *Game) attackCheckpointLocked() error {
	if g.attackRequirementsJudgedLocked() {
		return nil
	}
	g.RecomputeLayersIfStaleLocked()
	if e := g.attackRequirementsUnmetLocked(); e != nil {
		return e
	}
	g.markAttackDeclarationJudgedLocked()
	return nil
}
