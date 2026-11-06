package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// riders.go — #4978: the keyword actions a spell or an ability does
// beside its main effect, priced (ADR 0126's amendment of 2026-11-05).
//
// Preordain declared `draws: 1` and nothing else, so the bot priced it
// as Opt's draw, and Claim the Precious as a Murder: none of scry,
// learn, discover, amass or "the Ring tempts you" was a purpose amount.
// The catalog now declares each (`scry`, `learn`, `discover`, `amass`,
// `ring_tempts`), and riderValue prices them, added to whatever else the
// slot is worth wherever a purpose is read: a cast, a permanent's
// enters effect, an activation, a mode. A rider is always ADDED: inside
// purposeValue when the slot declares something else priced, and on top
// of the mana-value proxy, the body or ActivateBase when it does not
// (unpricedRiderValue).
//
//	riderValue = Hand × ScryWeight × min(scry, ScryCap)
//	           + Hand × LearnWeight × learn
//	           + Σ per discover (Hand + DiscoverPerMana × N)
//	           + the Army an amass grows or makes
//	           + RingTempt per tempting (half with no creature to bear it)
//
// Not priced: what the scry finds (the bot cannot know it), whether the
// library still holds a nonland card a discover can find, and how far
// along the Ring's levels the bot already is (CR 701.54c): the fifth
// tempting is priced like the first.

// riderSet is the riders a cast or an activation declares, summed over
// the slots it names.
type riderSet struct {
	scry, learn, amass, ringTempts int
	// discovers is how many discovers, discoverMana the sum of their N.
	discovers, discoverMana int
}

// add folds one declared purpose's riders in.
func (r *riderSet) add(p *protocol.PurposeView) {
	r.scry += p.Scry
	r.learn += p.Learn
	r.amass += p.Amass
	r.ringTempts += p.RingTempts
	if p.Discover > 0 {
		r.discovers++
		r.discoverMana += p.Discover
	}
}

// any reports whether a rider is declared.
func (r riderSet) any() bool {
	return r.scry != 0 || r.learn != 0 || r.amass != 0 || r.ringTempts != 0 || r.discovers != 0
}

// riderValue is what the declared riders are worth to the bot. `self`
// is the card being cast, if any: a creature spell whose enters effect
// tempts the bot is itself a creature to bear the Ring.
func (p *Policy) riderValue(st *state, r riderSet, self *protocol.CardView) float64 {
	if !p.cfg.PriceRiders || !r.any() {
		return 0
	}
	hand := st.w.Hand
	var v float64
	if r.scry > 0 {
		n := r.scry
		if p.cfg.ScryCap > 0 && n > p.cfg.ScryCap {
			n = p.cfg.ScryCap
		}
		v += hand * p.cfg.ScryWeight * float64(n)
	}
	v += hand * p.cfg.LearnWeight * float64(r.learn)
	// CR 701.57a: a discover is a card, cast for free when its mana
	// value is N or less, so each point of N is mana the bot keeps.
	v += hand*float64(r.discovers) + p.cfg.DiscoverPerMana*float64(r.discoverMana)
	v += st.amassValue(r.amass)
	if r.ringTempts > 0 {
		t := p.cfg.RingTempt * float64(r.ringTempts)
		if !st.controlsCreature() && (self == nil || !isCreature(self)) {
			t /= 2
		}
		v += t
	}
	return v
}

// unpricedRiderValue is riderValue for a purpose that declares nothing
// else this Config prices, and zero for one that does, which
// purposeValue has already priced with its riders. A rider never makes
// a slot "priced" on its own (purposePriced does not read it): Claim
// the Precious declares only its tempting, and pricing the spell by that
// alone would drop the mana-value proxy a removal spell is priced by. So
// the rider is added to whatever the slot is otherwise worth: the
// proxy, the body, ActivateBase.
func (p *Policy) unpricedRiderValue(st *state, ps purposeSet, self *protocol.CardView) float64 {
	if p.purposePriced(ps) {
		return 0
	}
	return p.riderValue(st, ps.riders, self)
}

// amassValue is what "amass N" adds to the bot's board (CR 701.47a): N
// +1/+1 counters on the Army it controls that gains the most from them,
// or a new N/N Army token, summoning-sick, when it controls none.
func (st *state) amassValue(n int) float64 {
	if n <= 0 {
		return 0
	}
	best, found := 0.0, false
	for _, c := range st.bf {
		if c.Controller != st.me || !isCreature(c) || !isType(c, "army") {
			continue
		}
		grown := *c
		grown.Power += n
		grown.Toughness += n
		if g := st.w.CreatureValue(&grown) - st.w.CreatureValue(c); !found || g > best {
			best, found = g, true
		}
	}
	if found {
		return best
	}
	army := protocol.CardView{TypeLine: "Token Creature — Army", Power: n, Toughness: n, IsToken: true, SummoningSick: true}
	return st.w.CreatureValue(&army)
}

// controlsCreature reports whether the bot controls a creature.
func (st *state) controlsCreature() bool {
	for _, c := range st.bf {
		if c.Controller == st.me && isCreature(c) {
			return true
		}
	}
	return false
}
