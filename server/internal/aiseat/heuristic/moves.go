package heuristic

import (
	"strconv"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// moves.go prices one move against doing nothing.
//
// ## Why this is not "Δscore on a cloned game"
//
// The S31 checklist says "move selection by Δscore on a cloned
// game". It cannot be done that way and keep ADR 0033 §3, and the
// guarantee wins. Cloning a game means holding a *game.Game; a
// Policy is handed a protocol.GameView and a []legal.Move and
// nothing else, precisely so that a policy CANNOT read a hand it is
// not entitled to. Reaching for a clone would mean importing
// internal/game into a policy package, which the import test in this
// directory fails the build over.
//
// So the delta is ESTIMATED rather than simulated: each move is
// priced in the same units score.go uses, by asking what the move
// does to the visible board. A land is a mana source; a creature
// spell is that creature minus the card it cost; a spell pointed at
// an opponent's creature is that creature's value, discounted by the
// confidence that a spell which may legally target an opponent's
// creature is removal rather than a pump. Passing is the zero.
//
// The honest limitation: with no oracle text on the wire for a
// single-faced card (CardView carries OracleText only inside Faces),
// the policy reads a spell's INTENT from what it is allowed to
// target and what it costs, not from what it says. That is the same
// information a player has when an opponent casts a face-down card,
// and it is enough for the Forge-grade bar this sprint sets. An
// oracle-text channel — ADR 0033's Input.Oracle — is what would
// close the gap, and it lands with the model tiers, not here.

// valueOf prices a single move. Positive is better than passing;
// exactly zero is "same as passing".
func (p *Policy) valueOf(st *state, m legal.Move) (float64, string) {
	switch m.Kind {
	case legal.KindPass:
		return 0, "pass"

	case legal.KindLand:
		// A land drop is free, once a turn, and almost never wrong.
		// It is priced above every ordinary cast so the bot plays its
		// land BEFORE it spends the turn's mana, which is the one
		// sequencing rule that matters at this level.
		return p.cfg.LandValue, "land drop"

	case legal.KindCast:
		return p.valueOfCast(st, m)

	case legal.KindActivate:
		cp := decode[activateParams](m.Params)
		v := p.cfg.ActivateBase
		v += st.targetsValue(p.cfg, cp.Targets)
		for _, id := range cp.SacrificeIDs {
			if c := st.bf[id]; c != nil {
				v -= st.w.permanentValue(c)
			}
		}
		if src := st.bf[cp.SourceCardID]; src != nil && isCreature(src) && !src.Tapped {
			// Tapping a creature for an ability costs us a blocker.
			v -= 0.3
		}
		return v, "activate"

	case legal.KindMana:
		// Casts already carry auto_tap, so floating mana buys the bot
		// nothing and empties at the next step boundary (CR 106.4).
		// Priced negative so it is only ever reached when literally
		// nothing else is on offer.
		return p.cfg.ManaFloat, "float mana"

	case legal.KindAttack, legal.KindBlock:
		// Combat is planned in combat.go. If a declaration reaches the
		// general scorer at all, the planner has already decided it
		// does not want one.
		return -1, "combat handled elsewhere"

	case legal.KindChoice:
		return p.valueOfChoice(st, m)
	}
	return 0, "unknown move kind"
}

func (p *Policy) valueOfCast(st *state, m legal.Move) (float64, string) {
	cp := decode[castParams](m.Params)
	card := st.mine[cp.InstanceID]
	var v float64
	reason := "cast"
	if card != nil {
		switch {
		case isCreature(card) || isPermanentSpell(card):
			// What the permanent will be worth once it resolves. It
			// arrives summoning-sick and untapped; permanentValue
			// reads SummoningSick off the hand card, which is false
			// there, so discount a creature explicitly.
			pv := st.w.permanentValue(card)
			if isCreature(card) {
				pv *= st.w.SickCreature
			}
			v += pv
			reason = "cast permanent"
		default:
			// An instant or sorcery: no body, so its value is its
			// targets plus a mana-value proxy for whatever it does
			// that the wire does not describe.
			v += p.cfg.SpellPerMana * float64(manaValue(card.ManaCost, cp.XValue))
			reason = "cast spell"
		}
		if card.IsCommander {
			v += p.cfg.CommanderBonus
		}
		if card.Unimplemented {
			// The engine will run none of this card's printed rules
			// (ADR 0037). It still costs a card.
			v -= p.cfg.SpellPerMana * float64(manaValue(card.ManaCost, cp.XValue))
			reason = "cast (unimplemented)"
		}
	}
	// The card leaves hand: one fewer resource.
	v -= st.w.Hand
	// Additional costs are paid out of the same pool of resources.
	v -= st.w.Hand * float64(len(cp.DiscardIDs))
	for _, id := range cp.SacrificeIDs {
		if c := st.bf[id]; c != nil {
			v -= st.w.permanentValue(c)
		}
	}
	v += st.targetsValue(p.cfg, cp.Targets)
	if cp.FromZone == "command" {
		// Each cast from the command zone makes the next one cost
		// {2} more (CR 903.8); the tax is already in the score, this
		// is the forward-looking half.
		v -= st.w.CommanderTax * 0.5
	}
	return v, reason
}

// isPermanentSpell reports whether a card in hand will become a
// permanent when it resolves.
func isPermanentSpell(c *protocol.CardView) bool {
	t := strings.ToLower(c.TypeLine)
	for _, k := range []string{"creature", "artifact", "enchantment", "planeswalker", "battle", "land"} {
		if strings.Contains(t, k) {
			return true
		}
	}
	return false
}

// targetsValue prices a target list. The sign convention is the
// whole heuristic: pointing a spell at someone else's permanent is
// worth what that permanent is worth, pointing it at your own is
// worth a little (it is presumably a pump or a protection), and
// pointing it at yourself is worth less than pointing it at anyone
// else.
func (st *state) targetsValue(cfg Config, targets []targetRef) float64 {
	var v float64
	for _, t := range targets {
		switch t.Kind {
		case "player":
			if t.ID == st.me {
				v -= cfg.SelfTargetPenalty
				continue
			}
			e := st.evals[t.ID]
			// Damage at a player is worth more the closer they are
			// to dead and the more of a problem they are.
			v += cfg.DamageToPlayer * st.leaderBoost(cfg, t.ID)
			switch {
			case e == nil:
			case e.Life <= cfg.FinishLife:
				// The bot cannot read how much damage the spell
				// deals — no oracle text reaches a policy — so this
				// is a bet that a spell it may point at a player who
				// is nearly dead will finish them. It is the right
				// bet: the alternative is what this replaced, a bot
				// that burned its last removal on a 7/7 while the
				// player at 1 life untapped and won. A missed kill
				// costs one card; a taken one ends a seat.
				v += cfg.LethalBonus
			case e.Life <= st.w.DangerLife:
				v += cfg.DamageToPlayer * 3
			}
		default:
			v += st.cardTargetValue(cfg, t.ID)
		}
	}
	return v
}

func (st *state) cardTargetValue(cfg Config, id string) float64 {
	if c := st.bf[id]; c != nil {
		if c.Controller == st.me {
			return cfg.OwnPermanentTarget
		}
		base := st.w.permanentValue(c)
		if isCreature(c) {
			base = st.w.CreatureValue(c)
		}
		return base * cfg.RemovalConfidence * st.leaderBoost(cfg, c.Controller)
	}
	if c := st.stack[id]; c != nil {
		if c.Controller == st.me {
			return -cfg.CounterValue * 2
		}
		return cfg.CounterValue
	}
	// A graveyard or exile target: recursion, reanimation, something
	// that puts a card back where we can use it.
	if c := st.graveyard[id]; c != nil {
		if c.Owner == st.me {
			return st.cardValue(cfg, c)
		}
		return cfg.OwnPermanentTarget
	}
	return 0
}

// leaderBoost scales a payoff by how much of a problem its owner is:
// the table's biggest threat is worth hitting more than the seat
// that has done nothing all game. Returns 1 for a non-leader.
func (st *state) leaderBoost(cfg Config, seat string) float64 {
	if seat == st.leader && seat != st.me {
		return cfg.LeaderBoost
	}
	return 1
}

// cardValue is "how much do I want this card", for cards that are
// not on the battlefield: in hand, in a graveyard, on top of a
// library. Used by the discard, sacrifice, search and scry branches.
func (st *state) cardValue(cfg Config, c *protocol.CardView) float64 {
	if c == nil {
		return 0
	}
	switch {
	case isLand(c):
		// Lands are wanted early and dead late.
		if st.myMana < cfg.LandsWanted {
			return st.w.ManaSource * 1.5
		}
		return st.w.ManaSource * 0.3
	case isCreature(c) || isPermanentSpell(c):
		v := st.w.permanentValue(c)
		if manaValue(c.ManaCost, 0) > st.myMana+1 {
			v *= 0.6 // can't cast it for a while
		}
		return v
	default:
		v := cfg.SpellPerMana * float64(manaValue(c.ManaCost, 0))
		if manaValue(c.ManaCost, 0) > st.myMana+1 {
			v *= 0.6
		}
		return v
	}
}

// manaValue parses a Scryfall mana-cost string into a converted
// cost. Generic digits add their value, {X} adds x, and every other
// symbol — coloured, hybrid, phyrexian, snow — counts as one.
func manaValue(cost string, x int) int {
	total := 0
	for {
		i := strings.IndexByte(cost, '{')
		if i < 0 {
			return total
		}
		j := strings.IndexByte(cost[i:], '}')
		if j < 0 {
			return total
		}
		sym := cost[i+1 : i+j]
		cost = cost[i+j+1:]
		switch {
		case sym == "X" || sym == "x":
			total += x
		case sym == "":
			// malformed; skip
		default:
			if n, err := strconv.Atoi(sym); err == nil {
				total += n
			} else {
				total++
			}
		}
	}
}

// colorSymbols counts each coloured pip across a set of mana costs,
// so the mana_pick branch can add the colour the bot's own hand
// actually wants.
func colorSymbols(cards []protocol.CardView) map[string]int {
	out := map[string]int{}
	for i := range cards {
		cost := cards[i].ManaCost
		for _, ch := range cost {
			switch ch {
			case 'W', 'U', 'B', 'R', 'G', 'C':
				out[string(ch)]++
			}
		}
	}
	return out
}
