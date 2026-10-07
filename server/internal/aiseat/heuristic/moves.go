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

// suicideValue is what a move that would pay the bot's last life is
// worth. It is eliminatedStrength — what score.go already prices a
// seat that is out of the game at — because that is precisely what
// such a move makes the bot, and the two numbers being the same one
// is the point rather than a coincidence.
const suicideValue = eliminatedStrength

// valueOf prices a single move. Positive is better than passing;
// exactly zero is "same as passing".
//
// Two halves: what the move BUYS, and what it CHARGES. The second is
// split out because it is the half the move list carries explicitly
// (legal.MoveCost, #74) and it applies to every kind — pricing it
// inside each branch of the switch would be four copies of the same
// arithmetic and a fifth one forgotten.
func (p *Policy) valueOf(st *state, m legal.Move) (float64, string) {
	v, reason := p.payoffOf(st, m)
	if m.Cost == nil {
		return v, reason
	}
	if m.Cost.PhyrexianLife > 0 && st.myLife()-m.Cost.Life < phyrexianLifeFloor {
		return phyrexianLifeDeclined, "Phyrexian life would take it below the floor"
	}
	c, refuse := p.costValue(st, st.bf[m.Source.String()], *m.Cost)
	if refuse {
		return suicideValue, "would pay its last life"
	}
	return v + c, reason
}

// costValue prices the cost components legal.Move declares, and
// reports whether the move has to be refused outright.
//
// THE REFUSAL BELONGS HERE, in the evaluation, rather than in a guard
// wrapped around the decision. score.go already prices a seat that is
// out of the game at eliminatedStrength, far below anything a live
// seat can reach; a move that pays the bot's last life makes the bot
// that seat, so answering with that number is the evaluation staying
// consistent with itself instead of a special case bolted on beside
// it. Everything downstream then needs no help: decideGeneral's
// threshold rejects it, the "no pass on offer" fallback rejects it
// (that one only takes a positive), and Rank sorts it last WITH its
// reason attached — which a guard that filtered the move out of the
// list would have hidden.
//
// It also does not fight concede.go. Declining to kill yourself this
// instant is not a judgement that the position is lost; that is the
// concede heuristic's business, it is measured over three turns, and
// nothing here touches it.
func (p *Policy) costValue(st *state, src *protocol.CardView, c legal.MoveCost) (float64, bool) {
	var v float64
	if c.Life > 0 {
		life := 0
		if st.myEval != nil {
			life = st.myEval.Life
		}
		if life-c.Life < p.cfg.LifeFloor {
			return 0, true
		}
		// The payoff proxy less the real price. The proxy is linear
		// in the life paid and the price is quadratic near death, so
		// one ability is a profit at 40 and a refusal at 9 without a
		// second rule saying so.
		//
		// #1677: the proxy is withheld from the life that pays a
		// Phyrexian symbol. That life buys the same spell the mana
		// would have — nothing more — so it is a pure price, and the
		// all-mana payment of the same announcement always scores
		// above it. Proxied, the bot would pay 4 life for a Dismember
		// with three Swamps untapped because the arithmetic called it
		// a profit.
		v += p.cfg.LifePayoff*float64(c.Life-c.PhyrexianLife) - st.w.LifeCostValue(life, c.Life)
	}
	if c.Loyalty != 0 {
		// Loyalty counters are board value the evaluation already
		// counts, so a +1 is a small gain and a −3 a real price.
		v += st.w.Loyalty * float64(c.Loyalty)
		if c.Loyalty < 0 && src != nil && src.Counters["loyalty"]+c.Loyalty <= 0 {
			// The last counter: the planeswalker dies to CR 704.5i
			// and the ability costs the whole permanent. The bot
			// cannot read what an ultimate does, so this is the only
			// half of that trade it can see — which is the right way
			// round, because an ultimate it cannot evaluate is not an
			// ultimate it should be firing.
			v -= st.permanentValue(src)
		}
	}
	for _, cp := range c.Counters {
		// #625: counters a cost removes, usually from a permanent
		// other than the source — Heart of Kiran's crew paid with a
		// planeswalker's loyalty. Priced against THAT permanent.
		v -= st.w.counterRemovalValue(st.bf[cp.CardID.String()], cp.Counter, cp.N)
	}
	// #1600: "Discard your hand" (Lion's Eye Diamond, Null Brooch,
	// Slate of Ancestry). The params name none of the cards, so the
	// count is the whole price, at what a card in hand is worth to the
	// evaluation — the weight a cast already pays for leaving the hand.
	// An empty hand costs nothing, which is how the cards are played.
	if c.Hand > 0 {
		v -= st.w.Hand * float64(c.Hand)
	}
	return v, false
}

// phyrexianLifeFloor is the life total the policy will not pay
// Phyrexian symbols below (CR 107.4f, #1677): a move whose
// legal.MoveCost.PhyrexianLife is non-zero is declined when paying its
// whole Life would leave the seat under this number.
//
// Ten, the default Weights.DangerLife — where the quadratic danger
// term starts to bite. Above it, 2 life for a symbol the board cannot
// pay is a cheap way to cast a spell the seat would otherwise hold;
// below it the life total is what the seat is racing with, and a
// Dismember that costs 4 of the last 12 is a spell that should wait
// for the mana. A constant rather than a Config field because it is a
// rule about this one cost shape, not a weight a tier tunes.
//
// It is a floor on PAYING life, not a reason to prefer it: above the
// floor the all-mana payment still wins wherever the enumerator
// offers one, because Phyrexian life gets no payoff proxy (costValue).
const phyrexianLifeFloor = 10

// phyrexianLifeDeclined is what a Phyrexian life payment under
// phyrexianLifeFloor is worth: below passing, so the policy never
// takes it, without the "would pay its last life" reason the
// CR 704.5a refusal carries — the seat would survive, it just should
// not spend that much.
const phyrexianLifeDeclined = -1.0

// myLife is the seat's own life total, or 0 when the view does not
// carry an evaluation for it.
func (st *state) myLife() int {
	if st.myEval == nil {
		return 0
	}
	return st.myEval.Life
}

// genericCounterValue is what one counter of a kind the evaluation
// does not otherwise read (gold, charge, stun, …) is worth to lose.
// Small and positive: a counter that exists is usually there to be
// spent, which is exactly what a cost that removes it does.
const genericCounterValue = 0.25

// stationPowerPayoff is what one point of a tapped creature's power is
// worth to an ability whose cost taps it (#759) — station's charge
// counters, one per point (CR 702.184a). Small: a tenth of what the
// same point is worth swinging (Weights.Power), so the attack always
// wins before combat and the counters win after it.
const stationPowerPayoff = 0.1

// zeroXActivation prices an activation of an {X} ability announced at
// X=0 (#810). Below passing, which scores 0, because the X-sized half
// of what the ability does is nothing at X=0 — and a repeatable
// no-op the policy keeps picking is a loop the game does not let run
// (CR 732.2a).
const zeroXActivation = -1.0

// counterRemovalValue prices removing n counters of `kind` from `c`,
// as a positive cost.
//
// The two kinds the evaluation already counts are priced in its own
// units, and both carry the cliff that matters: the LAST loyalty
// counter is the whole planeswalker (CR 704.5i), and the last point of
// toughness a +1/+1 counter was holding up is the whole creature
// (CR 704.5f). That cliff is what stops a bot crewing Heart of Kiran
// by killing a 1-loyalty walker for a Vehicle it had no plan for — the
// #74 lesson, one cost component over. A -1/-1 counter removed is a
// gain. A permanent the policy cannot see (nil) is priced at the
// generic rate rather than as free.
func (w Weights) counterRemovalValue(c *protocol.CardView, kind string, n int) float64 {
	if n <= 0 {
		return 0
	}
	switch kind {
	case "loyalty":
		v := w.Loyalty * float64(n)
		if c != nil && isType(c, "planeswalker") && c.Counters["loyalty"]-n <= 0 {
			v += w.permanentValue(c)
		}
		return v
	case "+1/+1":
		if c != nil && isCreature(c) && c.Toughness-n <= 0 {
			return w.permanentValue(c)
		}
		return (w.Power + w.Toughness) * float64(n)
	case "-1/-1":
		return -(w.Power + w.Toughness) * float64(n)
	}
	return genericCounterValue * float64(n)
}

// payoffOf prices what a move buys, before its declared cost.
func (p *Policy) payoffOf(st *state, m legal.Move) (float64, string) {
	switch m.Kind {
	case legal.KindPass:
		return 0, "pass"

	case legal.KindFinishBlocks:
		// #1501: a defender's "done declaring blockers", priced as
		// the pass it stands in for.
		return 0, "done blocking"

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
		// #810, belt and braces. The enumerator no longer offers an
		// {X} ability at X=0 when X is the whole of what it does, so
		// this should not be reachable for such a card — but the
		// policy is the other half of the pair that kept a table
		// spinning, and a move that buys nothing has to rank below
		// passing wherever it comes from. A card with a fixed rider
		// is still offered at X=0 and still lands here; the rider is
		// worth having, so the score is a small negative rather than
		// a refusal, and a bot takes it only when there is nothing
		// better on the list.
		if cp.XValue == 0 && st.abilityDemandsX(cp.SourceCardID, cp.AbilityIndex) {
			return zeroXActivation, "activate for X=0"
		}
		v := p.cfg.ActivateBase
		reason := "activate"
		// ADR 0106 §1 decision 8 (owner decision 2): an "Any player may
		// activate this ability" row on a permanent somebody else
		// controls is worth only what the row DECLARES it buys this
		// seat. The generic ActivateBase is positive, and an opponent's
		// Flailing Ogre pumped or Feral Hydra grown by this seat's mana
		// is a gift; so a row with no purpose is never chosen.
		src := st.bf[cp.SourceCardID]
		across := src != nil && src.Controller != "" && src.Controller != st.me
		if across {
			pv, ok := st.anyPlayerPurposeValue(src, cp.AbilityIndex)
			if !ok {
				return anyPlayerDeclined, "another player's ability with no declared purpose"
			}
			v, reason = pv, "activate another player's ability"
		}
		// ADR 0126 §6: a row of the bot's own that declares what it
		// does — a loot, a land search, a sweep — is priced by that, in
		// place of the flat ActivateBase, and a row that sacrifices its
		// own source pays for the source. A row with no purpose keeps
		// ActivateBase (owner decision 6).
		purposed := false
		if !across {
			var ps purposeSet
			ps.add(rowPurpose(src, cp.AbilityIndex))
			if p.purposePriced(ps) {
				v, reason, purposed = p.purposeValue(st, ps, cp.XValue, nil, false), "activate (declared purpose)", true
				// A row that taps a creature pays for it below, by when the
				// tap happens (tapCreatureCost): a loot before combat
				// costs the attack it replaces.
				if row := rowAt(src, cp.AbilityIndex); row != nil && row.SacrificeSelf {
					v -= st.permanentValue(src)
				}
			}
		}
		v += st.targetsValue(p.cfg, cp.Targets)
		// ADR 0126 §7: priced by the chance the bot would have kept
		// each permanent, less its death payoffs (sacrifice.go).
		v -= p.sacrificeCost(st, cp.SacrificeIDs)
		// #1600: a permanent exiled to pay the cost leaves the board as
		// surely as a sacrificed one, so it costs the same.
		for _, id := range cp.ExilePermanentIDs {
			if c := st.bf[id]; c != nil {
				v -= st.permanentValue(c)
			}
		}
		// #1297: an exile-N-cards cost spends real cards — a graveyard
		// card the seat might have recast, a card in hand. One price
		// for both, and it is the SAME one the enumerator ordered the
		// payment by (Options.OrderCostFuel), so the payment offered is
		// the one priced cheapest and the two cannot disagree — escape's
		// argument (#1013), one cost site over.
		for _, id := range cp.ExileIDs {
			v -= p.fuelValue(st, id)
		}
		// #2016: the cards a "Discard N cards" cost throws away, priced
		// as a cast's additional-cost discard is (discardCost: what each
		// is worth to the bot, a late land cheap, the last land dear),
		// less any discard payoff the bot's own permanents pay.
		v -= p.discardsCost(st, cp.SourceCardID, cp.DiscardIDs)
		// #2016: a keyword counter the source already has the keyword
		// for buys nothing (CR 122.1b), so the price paid is all loss.
		if src != nil && !across && redundantKeywordCounter(src, cp.AbilityIndex) {
			v += redundantCounterPenalty
			reason = "activate (redundant counter)"
		}
		// An {X} ability does more the bigger X is, and the
		// enumerator has already picked the largest X the seat can
		// actually pay (legal/abilities.go) — so the policy never
		// chooses X, it only prices the one on offer. Same
		// mana-value proxy the instant / sorcery branch of
		// valueOfCast uses, for the same reason: with no oracle text
		// on the wire, what the ability cost is the best available
		// signal for what it does.
		if cp.XValue > 0 && !purposed {
			v += p.cfg.SpellPerMana * float64(cp.XValue)
		}
		if src != nil && !across && isCreature(src) && !src.Tapped {
			// Tapping a creature for an ability costs us a blocker,
			// priced by when it happens (ADR 0126 §5) when the row's
			// cost really taps it. A row that does not tap keeps the
			// flat price it always had.
			if row := activatedRow(src, cp.AbilityIndex); row != nil && row.TapCost {
				v -= p.tapCreatureCost(st, src)
			} else {
				v -= tappedBlocker
			}
		}
		// Crewing taps creatures that would otherwise block, and the
		// enumerator names them, so the cost is visible here.
		for _, id := range cp.CrewIDs {
			if c := st.bf[id]; c != nil && !c.Tapped {
				v -= p.tapCreatureCost(st, c)
			}
		}
		// #1310: a waterbend payment taps artifacts and creatures the
		// enumerator names. A tapped creature is a blocker spent, like
		// crew; a non-creature artifact costs nothing this evaluator
		// can see, which is why the enumerator taps those first.
		for _, id := range cp.WaterbendIDs {
			if c := st.bf[id]; c != nil && !c.Tapped && isCreature(c) {
				v -= p.tapCreatureCost(st, c)
			}
		}
		// #759: a tap-another cost (station) taps a creature the
		// same way, so it costs the same blocker — and before combat
		// on our own turn it also costs that creature's ATTACK, which
		// is the trade station actually asks about. Priced as the
		// power it would have swung with, so a bot stations with the
		// creatures that did not attack (main phase 2) or could not
		// (summoning sick), rather than tapping its army down before
		// combat for a Spacecraft it cannot use this turn.
		//
		// The other side of the trade is what the ability buys, which
		// the policy cannot read. Station buys charge counters equal
		// to the tapped creature's power, and that power is the one
		// number on the move that measures it, so it is priced as a
		// small payoff per point (stationPowerPayoff). Without it the
		// blocker price alone sinks every activation below
		// PassThreshold and a bot never stations at all.
		for _, id := range cp.TapIDs {
			c := st.bf[id]
			if c == nil || c.Tapped {
				continue
			}
			if c.Power > 0 {
				v += stationPowerPayoff * float64(c.Power)
			}
			if p.cfg.TapByTiming {
				// ADR 0126 §5 generalised this branch's attack price to
				// every tapped creature; tapCreatureCost carries it.
				v -= p.tapCreatureCost(st, c)
				continue
			}
			v -= tappedBlocker
			if st.myTurn && st.step == "precombat_main" && !c.SummoningSick && c.Power > 0 {
				v -= st.w.Power * float64(c.Power)
			}
		}
		return v, reason

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

	case legal.KindSpecialAction:
		// CR 116.2. A special action puts nothing on the stack and
		// buys a cheaper or free cast on a later turn, which is why
		// it is priced as a flat positive rather than through
		// valueOfCast: there is no spell here to value, and the
		// payoff arrives on a turn this evaluator cannot see.
		//
		// The enumerator has already checked the timing and the
		// affordability, so anything that reaches here is a move the
		// engine will accept (#544).
		sp := decode[specialActionParams](m.Params)
		if sp.Kind == "unlock" {
			return p.valueOfUnlock(st, sp)
		}
		return p.cfg.SpecialActionValue, "special action: " + sp.Kind

	case legal.KindChoice:
		return p.valueOfChoice(st, m)
	}
	return 0, "unknown move kind"
}

func (p *Policy) valueOfCast(st *state, m legal.Move) (float64, string) {
	cp := decode[castParams](m.Params)
	card := st.castSource(cp.InstanceID)
	var v float64
	reason := "cast"
	if cp.AlternativeCost != "" {
		reason = "cast (" + cp.AlternativeCost + ")"
	}
	if card != nil {
		// What the card is worth once it resolves — the same function
		// the FUEL price reads, so "what an escape spends" and "what an
		// escape buys" cannot disagree about one card (#1013, fuel.go).
		// The targets are added below; a card being pitched points at
		// nothing, which is the one difference.
		// ADR 0126 §6: priced by the purpose of the cost or the modes
		// this cast names, so an overloaded Rift is a sweep and a
		// hard-cast one is not.
		ps := castPurpose(card, cp)
		v += p.resolvedValueFor(st, card, cp.XValue, ps, false, len(cp.Targets) > 0)
		// ADR 0126 §2: the ramp premium is a CAST price only. It is
		// what one more source is worth to a seat that is short of
		// mana now, which a card being pitched to a cost is not.
		if r := p.rampPremium(st, card); r > 0 {
			v += r
			reason = "cast mana source"
		}
		switch {
		case card.Unimplemented:
			reason = "cast (unimplemented)"
		case reason == "cast mana source":
		case isCreature(card) || isPermanentSpell(card):
			reason = "cast permanent"
		case p.cfg.PriceSweeps && len(ps.sweeps) > 0:
			reason = "cast wipe"
		case p.purposePriced(ps):
			reason = "cast spell (declared purpose)"
		default:
			reason = "cast spell"
		}
	}
	// A card leaves HAND: one fewer resource. #673 — a cast out of the
	// graveyard, exile or the top of the library costs no card in
	// hand, and charging one there was what made every flashback,
	// escape and impulse cast score below passing and never get taken.
	// What such a cast really spends is the card in the graveyard,
	// which since #1013 is the alternative cost's own price below
	// rather than a gap.
	if cp.FromZone == "" || cp.FromZone == "hand" || cp.FromZone == "command" {
		v -= st.w.Hand
	}
	// CR 601.2b's card half of a claimed alternative cost: Force of
	// Will's pitched blue card, Daze's Island off the board, escape's
	// five cards out of the graveyard.
	//
	// #1013: one price for all three, and it is the SAME one the
	// enumerator sorted the payments by — so the payment the bot is
	// offered first is the payment it then prices as cheapest, and the
	// two cannot disagree. A graveyard card used to be worth nothing
	// here, which made an escape look free.
	for _, id := range cp.AltCostIDs {
		v -= p.fuelValue(st, id)
	}
	// ADR 0100: delve's exiles are graveyard fuel too, priced by the
	// same function the enumerator ordered the delve pool by, so the
	// fewest-cards payment and the full-budget one are compared on what
	// each actually spends.
	for _, id := range cp.DelveIDs {
		v -= p.fuelValue(st, id)
	}
	// Additional costs are paid out of the same pool of resources.
	// ADR 0126 §7: a discarded card costs what IT is worth to the bot,
	// so the enumerator's one payment per discard combination lets the
	// bot pitch the spare land rather than the last one.
	if p.cfg.DiscardCostByCard {
		for _, id := range cp.DiscardIDs {
			v -= p.discardCost(st, id, cp.InstanceID, cp.DiscardIDs)
		}
	} else {
		v -= st.w.Hand * float64(len(cp.DiscardIDs))
	}
	// ADR 0126's amendment of 2026-10-06: a discard the bot's own
	// permanents pay for (Mary Read's Treasure for an Island) costs
	// that much less, so the enumerator's cheapest payment is the one
	// that triggers them.
	for _, id := range cp.DiscardIDs {
		v += st.discardPayoff(p.cfg, st.mine[id])
	}
	v -= p.sacrificeCost(st, cp.SacrificeIDs)
	v += st.targetsValue(p.cfg, cp.Targets)
	if cp.FromZone == "command" {
		// Each cast from the command zone makes the next one cost
		// {2} more (CR 903.8); the tax is already in the score, this
		// is the forward-looking half.
		v -= st.w.CommanderTax * 0.5
	}
	return v, reason
}

// rampPremium is ADR 0126 §2's cast-time premium for a new repeatable
// mana source: RampPerMana for each mana it makes, up to how short the
// bot is of casting what it holds.
//
//	want    = the largest mana value among the other cards in hand and
//	          the commander in the command zone (with its tax), capped
//	          at RampWantCap
//	sources = the mana the bot's own sources make, tapped or not
//	deficit = max(0, want − sources)
//
// The card being cast is left out of `want`: once it resolves the bot
// no longer holds it, so a Signet in a hand of one-drops closes no gap.
// The premium needs no turn decay, because sources grow and the
// deficit closes on its own. A land is never cast, and a one-shot
// source (a Treasure, a ritual) is not repeatable, so neither gets it.
func (p *Policy) rampPremium(st *state, card *protocol.CardView) float64 {
	if card == nil || isLand(card) {
		return 0
	}
	return p.rampFor(st, card, repeatableMana(card))
}

// rampFor is the ramp premium for `amount` more mana a turn, with the
// card `card` (nil for none) left out of `want`: a new mana source's,
// or the lands a ramp spell puts onto the battlefield (ADR 0126 §6,
// purposeValue).
func (p *Policy) rampFor(st *state, card *protocol.CardView, amount int) float64 {
	if p.cfg.RampPerMana == 0 || amount <= 0 || st.seat == nil {
		return 0
	}
	want := 0
	consider := func(c *protocol.CardView, tax int) {
		if (card != nil && c.InstanceID == card.InstanceID) || isLand(c) {
			return
		}
		if mv := manaValue(c.ManaCost, 0) + tax; mv > want {
			want = mv
		}
	}
	for i := range st.seat.Hand.Cards {
		consider(&st.seat.Hand.Cards[i], 0)
	}
	for i := range st.seat.Command.Cards {
		c := &st.seat.Command.Cards[i]
		consider(c, 2*st.seat.CommanderCasts[c.InstanceID])
	}
	if want > p.cfg.RampWantCap {
		want = p.cfg.RampWantCap
	}
	sources := 0
	for _, c := range st.bf {
		if c.Controller == st.me {
			sources += repeatableMana(c)
		}
	}
	deficit := want - sources
	if deficit <= 0 {
		return 0
	}
	return p.cfg.RampPerMana * float64(min(amount, deficit))
}

// castSource finds the card a cast move names, in any zone a cast can
// come out of: the bot's hand or command zone, any graveyard, the
// shared exile pile, and the top of the bot's own library (#673, CR
// 401.5).
//
// Before the enumerator walked those zones this could only ever be a
// hand card, so `st.mine` was the whole lookup. Leaving it that way
// once the moves arrived would have priced every flashback, escape,
// foretold and impulse cast as an unknown card — which is not a
// neutral answer: it scores below passing, and a bot offered a
// flashback would decline it forever.
//
// Nil when the card is in a zone this seat may not see, which is a
// real answer rather than a bug: the move is still offered and still
// priced, just without the body's value.
func (st *state) castSource(id string) *protocol.CardView {
	if c := st.mine[id]; c != nil {
		return c
	}
	if c := st.graveyard[id]; c != nil {
		return c
	}
	if c := st.exile[id]; c != nil {
		return c
	}
	if st.seat != nil {
		for i := range st.seat.Library.Cards {
			if st.seat.Library.Cards[i].InstanceID == id {
				return &st.seat.Library.Cards[i]
			}
		}
	}
	return nil
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

// abilityDemandsX reports whether the named ability of the named
// permanent carries an {X} in its mana component (#810).
//
// Read off the filtered view's own `demands_x`, which the server
// derives from the ability's cost string and ships so the client's X
// picker does not have to re-parse it (protocol.ActivatedAbilityView).
// A policy may not import internal/game (ADR 0033 §3), and this is
// what that rule looks like in practice: the fact is already on the
// wire, so the policy reads it rather than reaching for the engine.
// anyPlayerDeclined is what an "Any player may activate" row on
// another player's permanent is worth when the catalog declares no
// purpose for it (ADR 0106 §1 decision 8): below passing, so the
// policy never takes it.
const anyPlayerDeclined = -1.0

// redundantCounterPenalty sinks "put an indestructible counter on this"
// below passing when the source is already indestructible (#2016).
const redundantCounterPenalty = -1.0

// anyPlayerPurposeValue prices an "Any player may activate this
// ability" row on a permanent this seat does not control, from the
// purpose the catalog declared on it (ADR 0106 §1 decision 8, owner
// decision 2): the cards this seat draws, plus the life the
// permanent's controller loses — counted only while that controller is
// an opponent still in the game whose life can matter. Reports false
// when the row declares no purpose, or is not an any-player row at all,
// and then the move is never chosen.
//
// The purpose comes off the wire row (`purpose`), the only place a
// policy may read catalog data from (ADR 0033 §3).
func (st *state) anyPlayerPurposeValue(src *protocol.CardView, index int) (float64, bool) {
	var row *protocol.ActivatedAbilityView
	for i := range src.ActivatedAbilities {
		if src.ActivatedAbilities[i].Index == index {
			row = &src.ActivatedAbilities[i]
			break
		}
	}
	if row == nil || !row.AnyPlayer || row.Purpose == nil {
		return 0, false
	}
	v := st.w.Hand * float64(row.Purpose.Draws)
	if n := row.Purpose.ControllerLosesLife; n > 0 {
		if e := st.evals[src.Controller]; e != nil && !e.Eliminated && !e.CantLoseLife {
			v += (st.w.Life + st.w.ThreatLifeRev) * float64(n)
		}
	}
	return v, v > 0
}

func (st *state) abilityDemandsX(sourceID string, index int) bool {
	c := st.bf[sourceID]
	if c == nil || index < 0 || index >= len(c.ActivatedAbilities) {
		return false
	}
	return c.ActivatedAbilities[index].DemandsX
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
			case e.CantLoseLife:
				// ADR 0057 Decision 6: nothing finishes a seat that
				// can't lose to its life total, so no finishing bonus
				// and no danger premium — just the base damage value.
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
		base := st.permanentValue(c)
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

// manaValue parses a Scryfall mana-cost string into a mana value
// (CR 202.3). Generic digits add their value, {X} adds x, a
// monocoloured hybrid {2/W} adds its larger component, 2 (CR 202.3f),
// and every other symbol — coloured, two-colour hybrid, Phyrexian,
// snow — counts as one.
//
// This is a second copy of game.ParsedCost.ManaValue, and it has to
// be: a policy package may not import internal/game (ADR 0033 §3,
// imports_test.go), and the engine's parser rejects the joined
// "{1}{R} // {1}{U}" split-card cost this loop reads as the sum of
// its halves. TestManaValueMatchesTheRules pins it to the same table
// the engine's test uses; keep the two in step.
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
			} else if l, _, ok := strings.Cut(sym, "/"); ok {
				// Hybrid: the larger component (CR 202.3f). Only a
				// monocoloured {2/W} has a numeric half; {W/U},
				// {C/W} and Phyrexian {W/P} are all one.
				if n, err := strconv.Atoi(l); err == nil && n > 1 {
					total += n
				} else {
					total++
				}
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

// valueOfUnlock prices the CR 709.5e unlock special action (ADR 0103,
// owner decision 5) like casting a spell of the unlocked door's mana
// value: unlocking a door is paying that half's cost for that half's
// text, which is what casting it would have bought. An UNIMPLEMENTED
// Room's door does nothing the engine runs, so it falls back to the
// flat special-action value rather than being priced as a real spell.
func (p *Policy) valueOfUnlock(st *state, sp specialActionParams) (float64, string) {
	c := st.bf[sp.CardID]
	face := 0
	if sp.Door == "right" {
		face = 1
	}
	if c == nil || face >= len(c.Faces) || c.Unimplemented {
		return p.cfg.SpecialActionValue, "unlock door"
	}
	return p.cfg.SpellPerMana * float64(manaValue(c.Faces[face].ManaCost, 0)), "unlock " + c.Faces[face].Name
}
