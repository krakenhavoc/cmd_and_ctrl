package heuristic

import (
	"fmt"
	"slices"
	"sort"
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
// a keyword. Prefix-matched, so "protection from red" answers to
// "protection" and "ward {2}" to "ward".
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
	if hasKeyword(a, "deathtouch") {
		return true
	}
	return a.Power >= effectiveToughness(b)
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

	through := 0
	var stoppable []int
	for _, a := range attackers {
		n := 0
		for _, b := range blockers {
			if couldBlock(st, def.ID, a, b) {
				n++
			}
		}
		if n == 0 {
			through += a.Power // nothing they have can stand in its way
			continue
		}
		stoppable = append(stoppable, a.Power)
	}
	// The defender spends its blockers on the biggest threats first;
	// whatever is left over connects.
	sort.Sort(sort.Reverse(sort.IntSlice(stoppable)))
	for i := len(blockers); i < len(stoppable); i++ {
		through += stoppable[i]
	}
	return through >= def.Life
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
	// Everything currently pointed at this seat, and whether it has
	// already been stopped.
	blocked := map[string]bool{}
	incoming := map[string]*protocol.CardView{}
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.AttackingTarget == st.me {
			incoming[c.InstanceID] = c
		}
		if c.BlockingTarget != "" {
			blocked[c.BlockingTarget] = true
		}
	}
	unblocked := 0
	for id, c := range incoming {
		if !blocked[id] {
			unblocked += c.Power
		}
	}
	life := StartingLife
	if st.myEval != nil {
		life = st.myEval.Life
	}
	desperate := life-unblocked <= p.cfg.BlockChumpLife

	best, bestVal, bestReason := -1, 0.0, ""
	for i := range moves {
		if moves[i].Kind != legal.KindBlock {
			continue
		}
		bp := decode[blockParams](moves[i].Params)
		atk, blk := st.bf[bp.Attacker], st.bf[bp.Blocker]
		if atk == nil || blk == nil {
			continue
		}
		v := 0.0
		reason := "trade"
		if !blocked[atk.InstanceID] {
			saved := float64(atk.Power) * st.w.MarginalLife(life)
			if desperate {
				saved = float64(atk.Power) * p.cfg.DesperateDamage
				reason = "chump to survive"
			}
			v += saved
		} else {
			// Ganging up only pays if it changes the outcome.
			reason = "gang block"
		}
		if kills(blk, atk) {
			v += st.w.CombatValue(atk)
		}
		if kills(atk, blk) {
			v -= st.w.CombatValue(blk)
			if hasKeyword(atk, "deathtouch") {
				// A deathtoucher eats whatever blocks it; do not
				// feed it the best creature on the board.
				v -= st.w.CombatValue(blk) * 0.25
			}
		}
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
