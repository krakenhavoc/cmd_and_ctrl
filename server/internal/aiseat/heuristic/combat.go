package heuristic

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// combat.go plans attacks and blocks.
//
// Both are enumerated PER CREATURE, not per assignment (ADR 0033 §1
// — attack subsets across four opponents are not enumerable), and the
// runner re-enumerates after every applied move. So the planner does
// not build a whole declaration: it picks the single best remaining
// declaration and is asked again. A creature that has already been
// declared is tapped (attackers) or carries a blocking_target
// (blockers) and drops out of the next enumeration by itself, which
// is what makes one-at-a-time composition converge.

// hasKeyword reports whether a card's effective ability list carries
// a keyword. Prefix-matched, so "ward {2}" answers to "ward".
//
// NOT the way to ask about protection. A prefix match answers "has
// some protection" and throws the quality away, which is the question
// nothing in combat actually wants — see protectedFrom below, which
// goes through the engine's one reader (#662).
func hasKeyword(c *protocol.CardView, kw string) bool {
	if c == nil {
		return false
	}
	for _, a := range c.Abilities {
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(a)), kw) {
			return true
		}
	}
	return false
}

// protectedFrom reports whether `defender` has protection from a
// source with `source`'s characteristics (CR 702.16e).
//
// It reads CardView.Protection — the qualities the ENGINE parsed,
// through its one closed-grammar reader — rather than prefix-matching
// the raw ability token. A policy package may not import
// internal/game (ADR 0033 §3), so the projection is how the one
// reader reaches the bot; a second parser here is exactly the drift
// #662 exists to prevent.
//
// The bot needs the quality, not just the keyword: an attack into a
// pro-red blocker trades nothing if the attacker is red, and a prefix
// match cannot tell that from a pro-black one.
//
// Block LEGALITY (CR 702.16f) is not asked here and must not be — the
// enumerator reaches game.BlockPairRefusalLocked, so a block the
// engine refuses is never offered in the first place (ADR 0045 §3).
// This is the DAMAGE half (CR 702.16e), which decides whether an
// offered exchange is worth making.
func protectedFrom(defender, source *protocol.CardView) bool {
	if defender == nil || source == nil {
		return false
	}
	for _, p := range defender.Protection {
		switch p.Kind {
		case "everything":
			return true
		case "color":
			if slices.Contains(source.Colors, p.Value) {
				return true
			}
		case "mana_value_at_most":
			// #2181: the bound against the source's printed cost. A
			// minimal edit; the engine owns the rule (CR 202.3).
			if n, err := strconv.Atoi(p.Value); err == nil && manaValue(source.ManaCost, 0) <= n {
				return true
			}
		case "ring_bearer":
			// #2145: the designation the view already carries.
			if source.RingBearer {
				return true
			}
		case "card_type", "subtype":
			// The view's type line is the EFFECTIVE one, rendered by
			// the projection from the same characteristics the engine
			// matched against, so a word match on it is the same
			// answer the server gives. A whole-word test, so "Demon"
			// does not match "Demonlord": the type line is
			// space-and-dash separated by construction.
			if typeLineWord(source.TypeLine, p.Value) {
				return true
			}
		}
	}
	return false
}

// typeLineWord reports whether a rendered type line carries `word` as
// a whole token, case-insensitively. Supertypes, types and subtypes
// alike — the caller already knows which it is asking about, and no
// word is both.
func typeLineWord(typeLine, word string) bool {
	if word == "" {
		return false
	}
	for _, tok := range strings.FieldsFunc(typeLine, func(r rune) bool {
		return r == ' ' || r == '\u2014' || r == '-'
	}) {
		if strings.EqualFold(tok, word) {
			return true
		}
	}
	return false
}

// effectiveToughness is what is left before a creature dies.
func effectiveToughness(c *protocol.CardView) int {
	return c.Toughness - c.DamageMarked
}

// kills reports whether attacker-or-blocker `a` destroys `b` in one
// combat exchange.
func kills(a, b *protocol.CardView) bool {
	if hasKeyword(b, "indestructible") {
		return false
	}
	if a.Power <= 0 {
		return false
	}
	// CR 702.16e: damage from a source with the quality is prevented,
	// so an exchange with a protected creature is not an exchange at
	// all. Deathtouch does not get round it — prevented damage is
	// never dealt (#662).
	if protectedFrom(b, a) {
		return false
	}
	if hasKeyword(a, "deathtouch") {
		return true
	}
	return a.Power >= effectiveToughness(b)
}

// blockMove reads one KindBlock move as "this attacker, blocked by
// these creatures": one pair for declare_blocker, the whole group for
// declare_blockers (#750). Returns a nil attacker for a move whose
// cards are not on the board the bot can see.
func (st *state) blockMove(m legal.Move) (*protocol.CardView, []*protocol.CardView) {
	var pairs []blockParams
	if m.Type == legal.TypeDeclareBlockers {
		pairs = decode[blocksParams](m.Params).Blocks
	} else {
		pairs = []blockParams{decode[blockParams](m.Params)}
	}
	if len(pairs) == 0 {
		return nil, nil
	}
	atk := st.bf[pairs[0].Attacker]
	if atk == nil {
		return nil, nil
	}
	blockers := make([]*protocol.CardView, 0, len(pairs))
	for _, bp := range pairs {
		blk := st.bf[bp.Blocker]
		// A group is all-or-nothing to the engine, so it is
		// all-or-nothing to the score too.
		if blk == nil || bp.Attacker != pairs[0].Attacker {
			return nil, nil
		}
		blockers = append(blockers, blk)
	}
	return atk, blockers
}

// losses is the blockers the attacker's damage kills. The attacker's
// controller divides its damage as it likes (CR 510.1a, 510.1c), so an
// estimate has to assume it spends that damage as badly for the
// defender as it can: the most combat value whose lethal damage adds up
// to no more than what it has to assign. A deathtoucher needs one point
// each (lethalFrom), a double striker divides its power twice, and an
// attacker the blockers' first strike destroys before it deals damage
// kills nothing (assignable, #1549).
//
// The most value, not the most bodies (#1548): a block is scored by
// what it adds to the group's losses, and an attacker that killed the
// cheapest first would let a defender "save" the Wurm blocking a Wurm
// by throwing an Ogre in beside it. Nobody assigns damage that way. It
// is a 0/1 knapsack over a few blockers and a few points of damage,
// solved exactly. A tie between two sets of equal value kills the
// earlier blockers, so a group with a new blocker added at the end
// loses the ones it was already losing when nothing is worse off.
func (st *state) losses(atk *protocol.CardView, blockers []*protocol.CardView) []*protocol.CardView {
	power := assignable(atk, blockers)
	if power <= 0 || len(blockers) == 0 {
		return nil
	}
	n := len(blockers)
	cost := make([]int, n)
	// best[i][c] is the most value killable among blockers[:i] with c
	// damage to assign.
	best := make([][]float64, n+1)
	best[0] = make([]float64, power+1)
	for i := 1; i <= n; i++ {
		cost[i-1] = lethalFrom(atk, blockers[i-1])
		best[i] = make([]float64, power+1)
		v := st.w.CombatValue(blockers[i-1])
		for c := 0; c <= power; c++ {
			best[i][c] = best[i-1][c]
			if k := cost[i-1]; k >= 0 && k <= c {
				if with := best[i-1][c-k] + v; with > best[i][c] {
					best[i][c] = with
				}
			}
		}
	}
	// Walk back from the last blocker: one is killed only when it
	// strictly adds value, so a tie keeps the earlier blockers.
	var out []*protocol.CardView
	c := power
	for i := n; i >= 1; i-- {
		if best[i][c] != best[i-1][c] {
			out = append(out, blockers[i-1])
			c -= cost[i-1]
		}
	}
	slices.Reverse(out)
	return out
}

// --- attacks -------------------------------------------------------

// decideAttack picks one attacker to declare, or reports that the bot
// is done attacking this turn.
func (p *Policy) decideAttack(st *state, moves []legal.Move) (aiseat.Decision, bool) {
	if len(st.opps) == 0 {
		return aiseat.Decision{}, false
	}
	focus := p.agg.pick(st.turn, st.opps)

	// Creatures still available to declare, in enumeration order.
	seen := map[string]bool{}
	var available []*protocol.CardView
	for i := range moves {
		if moves[i].Kind != legal.KindAttack {
			continue
		}
		id := decode[attackParams](moves[i].Params).Attacker
		if seen[id] {
			continue
		}
		seen[id] = true
		if c := st.bf[id]; c != nil {
			available = append(available, c)
		}
	}
	if len(available) == 0 {
		return aiseat.Decision{}, false
	}

	// How many bodies to keep home. A bot that alpha-strikes into
	// three opponents' untapped boards dies to the crack-back, and
	// that is the single most common way a naive multiplayer AI
	// loses a game it was winning.
	keepHome := 0
	if st.myEval != nil && st.myEval.Life < p.cfg.ReserveLife && p.opposingCreatures(st) > 0 {
		keepHome = p.cfg.AttackReserve
	}

	// Seats an all-in swing would actually kill. Nothing else the bot
	// could do this turn is worth more.
	push := map[string]bool{}
	for _, o := range st.opps {
		if p.lethalPush(st, o) {
			push[o.ID] = true
		}
	}

	// #1409: no kill this turn, but maybe one next turn that nothing
	// can stop in between. A committed race decides the whole attack —
	// who swings, at whom, and that everyone else stays home — because
	// its crack-back check counted on exactly that reserve (race.go).
	if len(push) == 0 {
		if plan := p.planRace(st, moves, focus); plan != nil {
			return p.raceAttack(moves, plan)
		}
		// #1548: no race inside two turns, but maybe one after a few
		// even trades that keep the edge (attrition.go). The plan is
		// sent exactly as a race is, reserve and all.
		if plan := p.planAttrition(st, moves, focus); plan != nil {
			return p.raceAttack(moves, plan)
		}
	}

	best, bestVal, bestReason := -1, 0.0, ""
	for i := range moves {
		if moves[i].Kind != legal.KindAttack {
			continue
		}
		ap := decode[attackParams](moves[i].Params)
		atk := st.bf[ap.Attacker]
		def := st.evals[ap.Target]
		if atk == nil || def == nil {
			continue
		}
		if keepHome > 0 && len(available) <= keepHome && !hasKeyword(atk, "vigilance") && !push[ap.Target] {
			continue
		}
		declared, declaredPower := declaredAgainst(st, ap.Target)
		v, reason := p.attackValue(st, atk, def, declared, declaredPower)
		// ADR 0080 / #1063: the CR 508.1a attack tax. The enumerator
		// has already dropped every attack this seat cannot pay for,
		// so what is left to decide is whether the attack is WORTH
		// the mana — a 1/1 into Ghostly Prison spends the turn's two
		// lands for two damage, and the bot should usually pass.
		//
		// Read off MoveCost.Mana rather than re-derived: a policy may
		// not import internal/game (ADR 0033 §3), and the tax is
		// nowhere else on the wire.
		if tax := attackTaxValue(moves[i]); tax > 0 {
			v -= float64(tax) * p.cfg.AttackTaxPenalty
			reason += fmt.Sprintf(" (pays %d)", tax)
		}
		if push[ap.Target] {
			v += p.cfg.LethalBonus
			reason = "all-in for the kill"
		}
		if ap.Target == focus {
			v += p.cfg.FocusBonus
			reason += " (focus)"
		}
		if best < 0 || v > bestVal {
			best, bestVal, bestReason = i, v, reason
		}
	}
	if best >= 0 && bestVal > p.cfg.PassThreshold {
		return aiseat.Decision{
			Index:  best,
			Reason: fmt.Sprintf("attack: %s (+%.2f)", bestReason, bestVal),
		}, true
	}
	return aiseat.Decision{}, false
}

// opposingCreatures counts the untapped creatures every opponent
// controls — the crack-back the bot is deciding whether to expose
// itself to.
func (p *Policy) opposingCreatures(st *state) int {
	n := 0
	for _, e := range st.opps {
		n += e.UntappedCreatures
	}
	return n
}

// declaredAgainst counts this seat's creatures already attacking a
// defender, and the damage they carry. It is what turns a string of
// one-at-a-time declarations into an alpha strike: once the attack
// outnumbers the blockers, every further attacker is unblockable, and
// a planner that cannot see that will sit behind an even board until
// somebody's library runs out.
func declaredAgainst(st *state, defender string) (count, power int) {
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller == st.me && c.AttackingTarget == defender {
			count++
			power += c.Power
		}
	}
	return count, power
}

// defenderBlockers lists a seat's untapped creatures — everything it
// could put in front of an attacker.
func defenderBlockers(st *state, seat string) []*protocol.CardView {
	var out []*protocol.CardView
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller == seat && isCreature(c) && !c.Tapped {
			out = append(out, c)
		}
	}
	return out
}

// lethalPush asks the question a creature-at-a-time planner cannot:
// if everything swung at once, would this seat survive it?
//
// Without it two low-life bots sit across an even board forever. Each
// individual attack is unprofitable — the defender blocks it, and at
// two life the damage a block saves them is worth more than any
// creature — so neither ever declares the first attacker, and a game
// that should end on turn 20 grinds to a library-out. The defender
// can only block as many attackers as it has creatures, and the
// overflow is what kills; that is a property of the whole attack, not
// of any one creature in it.
func (p *Policy) lethalPush(st *state, def *SeatEval) bool {
	if def.Life <= 0 {
		return false
	}
	// ADR 0057 Decision 6: damage can't finish a seat whose life
	// can't make it lose, so an all-in swing at it is never lethal.
	if def.CantLoseLife {
		return false
	}
	blockers := defenderBlockers(st, def.ID)
	// The swing is what is already committed to this seat plus
	// everything still able to join it.
	var attackers []*protocol.CardView
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller != st.me || !isCreature(c) {
			continue
		}
		committed := c.AttackingTarget == def.ID
		joinable := c.AttackingTarget == "" && !c.Tapped && !c.SummoningSick && !hasKeyword(c, "defender")
		if committed || joinable {
			attackers = append(attackers, c)
		}
	}

	return unblockedPower(st, def.ID, attackers, blockers) >= def.Life
}

// unblockedPower is the damage that connects when the defender blocks
// as well as it can: each blocker stops at most one attacker, and only
// an attacker it could legally block (couldBlock).
//
// The defender stops the biggest threats it is ABLE to stop. That is a
// matching, not a count, and #1261 is what counting got wrong: the old
// model let every untapped creature stand in front of any attacker, so
// 22 attackers into 22 defenders connected for nothing even when seven
// of the attackers were Drakes and five of the defenders could fly. A
// Bear cannot block a Drake. Two heuristic seats sat behind exactly
// that board, each with lethal on it, until the turn budget ran out.
//
// Greedy over attackers by power, keeping one only if the blocks
// chosen so far can be re-matched to make room for it (Kuhn's
// augmenting path). With the weight on the attacker side alone this
// is exact — the blockable sets form a transversal matroid, where
// greedy by weight is optimal — so the answer is the defender's true
// best, never a guess in either direction. Boards are a few dozen
// creatures a side, so the cubic worst case is nothing.
//
// Trample (#1504): a blocked trampler still connects for whatever its
// blockers do not absorb (CR 702.19b), and a matching that counts it
// as stopped never sees a one-Wurm edge — a Bear in front of a 7/7 is
// 5 to the face, every turn. When the swing has a trampler, the answer
// is the larger of two lower bounds: the matching's (every block holds
// in full, which a blocked trampler's overflow can only add to) and
// trampleBound's (the overflow, with the defender's blocks relaxed so
// that it can only be undercounted). Both are lower bounds on the
// defender's true best, so their maximum is too, and a swing without a
// trampler never computes the second.
//
// Still deliberately simple about the rest of combat: menace would
// only make the defender's job harder, and damage prevention is not
// modelled, so it errs toward "not lethal" wherever it errs.
func unblockedPower(st *state, defender string, attackers, blockers []*protocol.CardView) int {
	through, _ := matchBlocks(attackers, blockers, func(a, b *protocol.CardView) bool {
		return couldBlock(st, defender, a, b)
	})
	for _, a := range attackers {
		if hasKeyword(a, "trample") && a.Power > 0 {
			return max(through, trampleBound(st, defender, attackers, blockers))
		}
	}
	return through
}

// matchBlocks is unblockedPower's matching with the edge test as a
// parameter, so the two-turn race (race.go) can ask the same question
// about a narrower set of blocks — the ones a blocker survives, or the
// ones that hold a trampler in full. It returns the power that goes
// unblocked and, per attacker, the index of the blocker stopping it
// (-1 when none does).
func matchBlocks(attackers, blockers []*protocol.CardView, can func(a, b *protocol.CardView) bool) (int, []int) {
	edges := make([][]int, len(attackers))
	for ai, a := range attackers {
		for bi, b := range blockers {
			if can(a, b) {
				edges[ai] = append(edges[ai], bi)
			}
		}
	}
	blockerOf := make([]int, len(blockers)) // blocker index → attacker it stops, or -1
	for i := range blockerOf {
		blockerOf[i] = -1
	}
	var augment func(ai int, seen []bool) bool
	augment = func(ai int, seen []bool) bool {
		for _, bi := range edges[ai] {
			if seen[bi] {
				continue
			}
			seen[bi] = true
			if blockerOf[bi] < 0 || augment(blockerOf[bi], seen) {
				blockerOf[bi] = ai
				return true
			}
		}
		return false
	}
	through := 0
	for _, ai := range byPower(attackers) {
		if !augment(ai, make([]bool, len(blockers))) {
			through += attackers[ai].Power
		}
	}
	stoppedBy := make([]int, len(attackers))
	for i := range stoppedBy {
		stoppedBy[i] = -1
	}
	for bi, ai := range blockerOf {
		if ai >= 0 {
			stoppedBy[ai] = bi
		}
	}
	return through, stoppedBy
}

// byPower is the attackers' indexes, biggest threat first; stable, so
// equal powers keep board order and no answer depends on a sort's whim.
func byPower(attackers []*protocol.CardView) []int {
	order := make([]int, len(attackers))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool {
		return attackers[order[i]].Power > attackers[order[j]].Power
	})
	return order
}

// attackValue prices one attacker against one defending seat. The
// model is deliberately shallow: what does the defender's best
// single block do to this creature, and is the damage worth it.
func (p *Policy) attackValue(st *state, atk *protocol.CardView, def *SeatEval, declared, declaredPower int) (float64, string) {
	damage := float64(atk.Power) * p.cfg.DamageToOpponent
	if hasKeyword(atk, "lifelink") && st.myEval != nil {
		damage += float64(atk.Power) * st.w.MarginalLife(st.myEval.Life)
	}

	// The defender's untapped creatures that could legally block THIS
	// attacker are their answer. A tapped board, or a ground board
	// against a flier, is an open door.
	var blockers []*protocol.CardView
	for _, c := range defenderBlockers(st, def.ID) {
		if couldBlock(st, def.ID, atk, c) {
			blockers = append(blockers, c)
		}
	}
	// Every creature already attacking this seat is a blocker it has
	// to spend; what is left over is what can answer THIS attacker.
	free := len(blockers) - declared
	if free <= 0 {
		if declaredPower+atk.Power >= def.Life {
			return damage + p.cfg.LethalBonus, "lethal, unblockable"
		}
		if len(blockers) == 0 {
			return damage, "undefended"
		}
		return damage, "outnumbers their blockers"
	}

	// The defender blocks with whichever creature profits them most.
	worst := 0.0
	for _, b := range blockers {
		gain := 0.0
		if kills(b, atk) {
			gain += st.w.CombatValue(atk)
		}
		if kills(atk, b) {
			gain -= st.w.CombatValue(b)
		}
		// …and the damage the block saves them.
		gain += float64(atk.Power) * st.w.MarginalLife(def.Life)
		if gain > worst {
			worst = gain
		}
	}
	if worst <= 0 {
		// Nothing they can do about it profitably.
		return damage, "blockers can't profit"
	}
	return damage - worst, "blocked at a loss"
}

// attackTaxValue is what this attack move charges at CR 508.1a, as a
// mana value. Zero for every move at a table with no attack tax on it,
// which is the overwhelming majority (ADR 0080, #1063).
//
// The X is zero because an attack tax never carries {X}: a card whose
// price scales renders the count as a generic number before the cost
// string leaves the engine (Sphere of Safety's "{3}"), for exactly
// this reason — a declaration has nowhere to announce an X.
func attackTaxValue(mv legal.Move) int {
	if mv.Cost == nil || mv.Cost.Mana == "" {
		return 0
	}
	return manaValue(mv.Cost.Mana, 0)
}

// couldBlock is the evasion check the defender has to beat. The
// enumerator already applies the real CR 509.1b test to the blocks it
// OFFERS; this is the attacker's side of the same question, where
// there is no move list to read it off, so it re-derives the
// restrictions that matter at this catalog's level. It is an estimate
// (ADR 0045 addendum Decision 17): it only ranks moves the enumerator
// has already made legal, so being wrong costs a worse attack, never
// an illegal one.
//
// `defender` is the seat the attack would hit; landwalk reads that
// seat's lands off the public view. Fear, intimidate, shadow,
// horsemanship, skulk, and restriction tokens read the same effective
// public characteristics the engine projects.
func couldBlock(st *state, defender string, atk, blk *protocol.CardView) bool {
	if hasKeyword(atk, "flying") && !hasKeyword(blk, "flying") && !hasKeyword(blk, "reach") {
		return false
	}
	if hasRestriction(blk, "cant_block") || hasRestriction(atk, "cant_be_blocked") {
		return false
	}
	if landwalkBites(st, defender, atk) {
		return false
	}
	if hasKeyword(atk, "fear") && !isArtifact(blk) && !hasColor(blk, "B") {
		return false
	}
	if hasKeyword(atk, "intimidate") && !isArtifact(blk) && !sharesColor(atk, blk) {
		return false
	}
	if hasKeyword(atk, "shadow") != hasKeyword(blk, "shadow") {
		return false
	}
	if hasKeyword(atk, "horsemanship") && !hasKeyword(blk, "horsemanship") {
		return false
	}
	if hasKeyword(atk, "skulk") && powerForComparison(blk) > powerForComparison(atk) {
		return false
	}
	return true
}

func powerForComparison(c *protocol.CardView) int {
	if c.NegativePower < 0 {
		return c.NegativePower
	}
	return c.Power
}

// hasRestriction reads the server-projected restriction set. Restrictions
// are effects on a card, not keywords in its ability list.
func hasRestriction(c *protocol.CardView, want string) bool {
	return c != nil && slices.Contains(c.Restrictions, want)
}

func hasColor(c *protocol.CardView, want string) bool {
	return c != nil && slices.Contains(c.Colors, want)
}

func sharesColor(a, b *protocol.CardView) bool {
	if a == nil || b == nil {
		return false
	}
	for _, color := range a.Colors {
		if hasColor(b, color) {
			return true
		}
	}
	return false
}

func isArtifact(c *protocol.CardView) bool { return isType(c, "artifact") }

// landwalkLand is what one landwalk keyword asks of a land (CR
// 702.14c): a land subtype, or, for nonbasic landwalk, the absence of
// the basic supertype. The engine's own table is game/landwalk.go;
// this package may not import it (ADR 0033 §3), so the estimate keeps
// the six tokens it knows.
var landwalkLand = []struct {
	keyword  string
	subtype  string
	nonbasic bool
}{
	{"plainswalk", "plains", false},
	{"islandwalk", "island", false},
	{"swampwalk", "swamp", false},
	{"mountainwalk", "mountain", false},
	{"forestwalk", "forest", false},
	{"nonbasic landwalk", "", true},
}

// landwalkBites reports whether `atk` has a landwalk keyword that a
// land `defender` controls switches on, reading each land's effective
// type line from the view ("Basic Land — Island").
func landwalkBites(st *state, defender string, atk *protocol.CardView) bool {
	for _, lw := range landwalkLand {
		if !hasKeyword(atk, lw.keyword) {
			continue
		}
		for i := range st.view.Battlefield.Cards {
			c := &st.view.Battlefield.Cards[i]
			if c.Controller != defender || !isLand(c) {
				continue
			}
			left, right, _ := strings.Cut(strings.ToLower(c.TypeLine), "—")
			if lw.nonbasic {
				if !slices.Contains(strings.Fields(left), "basic") {
					return true
				}
				continue
			}
			if slices.Contains(strings.Fields(right), lw.subtype) {
				return true
			}
		}
	}
	return false
}

// --- blocks --------------------------------------------------------

// decideBlock picks one block to declare, or reports that the bot is
// done blocking.
func (p *Policy) decideBlock(st *state, moves []legal.Move) (aiseat.Decision, bool) {
	// Everything currently pointed at this seat, with whatever already
	// blocks it (#1706: a creature blocking several attackers is in
	// front of each).
	incoming := map[string]*protocol.CardView{}
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.AttackingTarget == st.me {
			incoming[c.InstanceID] = c
		}
	}
	// What still connects: every unblocked attacker's power, and every
	// blocked trampler's overflow (#1548) — a chumped Wurm is still 5
	// to the face, and a seat on 4 is as dead to it as to an unblocked
	// one.
	through := 0
	for id, c := range incoming {
		through += connects(c, blockersOn(st, id))
	}
	life := StartingLife
	if st.myEval != nil {
		life = st.myEval.Life
	}
	desperate := life-through <= p.cfg.BlockChumpLife

	best, bestVal, bestReason := -1, 0.0, ""
	for i := range moves {
		if moves[i].Kind != legal.KindBlock {
			continue
		}
		// #750: a block move is either one pair (declare_blocker) or
		// a whole group (declare_blockers) that is legal only
		// together, such as the two creatures a menace attacker
		// takes. Both are scored as ONE block of one attacker by the
		// creatures named, which is what they are.
		atk, adding := st.blockMove(moves[i])
		if atk == nil || len(adding) == 0 {
			continue
		}
		// #1548: the move adds to whatever already blocks the
		// attacker, and is scored as the whole group against it, less
		// what the group without it already did. A second Ogre on a
		// Wurm one Ogre is chumping turns the chump into a kill, and
		// the Wurm's damage still kills only one of them.
		have := blockersOn(st, atk.InstanceID)
		gang := append(append(make([]*protocol.CardView, 0, len(have)+len(adding)), have...), adding...)
		v := 0.0
		reason := "trade"
		if len(adding) > 1 {
			reason = "menace block"
		}
		if len(have) > 0 {
			// Ganging up only pays if it changes the outcome: a kill,
			// fewer losses, or less trample damage over the top.
			reason = "gang block"
		}
		if saved := connects(atk, have) - connects(atk, gang); saved > 0 {
			if desperate {
				v += float64(saved) * p.cfg.DesperateDamage
				if len(have) == 0 {
					reason = "chump to survive"
				}
			} else {
				v += float64(saved) * st.w.MarginalLife(life)
			}
		}
		if gangKills(atk, gang) && !gangKills(atk, have) {
			v += st.w.CombatValue(atk)
			if len(have) > 0 {
				reason = "gang block to kill"
			}
		}
		// The attacker assigns its damage among the whole group, so
		// the group loses whichever of them that damage can kill
		// (losses), and this block costs what that adds to the loss
		// the group without it was already taking.
		v -= st.lossValue(atk, gang) - st.lossValue(atk, have)
		if best < 0 || v > bestVal {
			best, bestVal, bestReason = i, v, reason
		}
	}
	if best >= 0 && bestVal > 0 {
		return aiseat.Decision{
			Index:  best,
			Reason: fmt.Sprintf("block: %s (+%.2f)", bestReason, bestVal),
		}, true
	}
	return aiseat.Decision{}, false
}
