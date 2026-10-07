package heuristic

import (
	"context"
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// choices.go answers the pending-choice windows. These are the
// windows a bot MUST answer — the engine refuses pass_priority while
// any choice is open — so unlike a priority window there is no
// "decline" here: one of the offered answers is going to happen.
//
// The choice's Kind comes off GameView.PendingChoices rather than
// being guessed from which payload fields are populated. That matters:
// "these card IDs" means discard the worst of them for one kind and
// take the best of them for another, and the sign is the whole
// decision.

// pending choice kinds, as they appear on the wire
// (game.PendingChoiceKind).
const (
	choiceDiscardFromHand     = "discard_from_hand"
	choiceRevealedHandPick    = "revealed_hand_pick"
	choiceMana                = "mana_pick"
	choiceReplacementOrder    = "replacement_order"
	choiceOptionalReplacement = "optional_replacement"
	choiceDamageAssignment    = "damage_assignment"
	choiceTriggerPrompt       = "trigger_prompt"
	choicePayUnless           = "pay_unless"
	choiceTriggerOrder        = "trigger_order"
	choicePickTarget          = "pick_target"
	choiceSacrifice           = "sacrifice_choice"
	choiceScry                = "scry"
	choicePutInLibrary        = "put_in_library"
	choiceSearchLibrary       = "search_library"
	choiceEntryPayLife        = "entry_pay_life"
	choiceConfirm             = "confirm"
	choiceChooseCards         = "choose_cards"
	choiceUntapChoice         = "untap_choice"
	choiceEntryRevealFromHand = "entry_reveal_from_hand"
	choiceEntryDiscard        = "entry_discard_from_hand"
	choiceEntrySacrifice      = "entry_sacrifice"
	choiceColor               = "choose_color"
	choiceCoinCall            = "coin_call"
	choiceEntryController     = "entry_controller"
	choiceMayCast             = "may_cast"
	choiceChooseSource        = "choose_source"
	choiceCommanderReturn     = "commander_return"
)

// damageSourceThreat ranks a choose_source candidate (ADR 0107 §6
// decision 4): below zero for the bot's own, above every permanent for an
// opponent's spell on the stack, and otherwise an opponent's power. Its own
// sources are ranked by power among themselves, still below any
// opponent's, so a prompt that offers only the bot's own (Desperate
// Gambit's "a source you control", ADR 0108 §3) picks its hardest hitter.
func (st *state) damageSourceThreat(id string, c *protocol.CardView) float64 {
	if c == nil {
		return 0
	}
	if c.Controller == st.me {
		return -1 + float64(c.Power)/1000
	}
	if _, onStack := st.stack[id]; onStack {
		return 100 + float64(c.Power)
	}
	return 1 + float64(c.Power)
}

// decideChoice takes the highest-valued answer. Ties go to the lowest
// index, which is the enumerator's own preference order.
func (p *Policy) decideChoice(ctx context.Context, st *state, moves []legal.Move) aiseat.Decision {
	best, bestVal, bestReason := 0, 0.0, "first offered answer"
	for i := range moves {
		if i%16 == 0 && ctx.Err() != nil {
			break
		}
		v, reason := p.valueOfChoice(st, moves[i])
		if i == 0 || v > bestVal {
			best, bestVal, bestReason = i, v, reason
		}
	}
	return aiseat.Decision{Index: best, Reason: bestReason}
}

func (p *Policy) valueOfChoice(st *state, m legal.Move) (float64, string) {
	// The cleanup-step discard is its own action type, not a pending
	// choice: pitch whatever the bot wants least.
	if m.Type == legal.TypeDiscardSelection {
		var v float64
		for _, id := range decode[discardSelectionParams](m.Params).CardIDs {
			c := st.mine[id]
			v -= st.cardValue(p.cfg, c) - st.discardPayoff(p.cfg, c)
		}
		return v, "discard to hand size"
	}

	cp := decode[choiceParams](m.Params)
	ch := st.choices[cp.ChoiceID]
	kind := ""
	opts := map[string]*protocol.CardView{}
	if ch != nil {
		kind = ch.Kind
		for i := range ch.Options {
			opts[ch.Options[i].InstanceID] = &ch.Options[i]
		}
	}
	lookup := func(id string) *protocol.CardView {
		if c := opts[id]; c != nil {
			return c
		}
		if c := st.mine[id]; c != nil {
			return c
		}
		if c := st.bf[id]; c != nil {
			return c
		}
		return st.graveyard[id]
	}

	switch kind {
	case choiceEntryController:
		return st.entryControllerValue(ch, cp.OptionIndex)
	case choiceEntryRiot:
		return st.riotValue(ch, cp.Apply)
	case choiceDamageAssignment:
		// The enumerator offers exactly one canonical split: the
		// prefix-lethal one a player makes almost every time.
		return 1, "canonical damage assignment"

	case choiceMana:
		// #742: a one-pick-N-mana choice adds a different amount per
		// colour (Nyx Lotus's devotion). The amount leads, so four {G}
		// beats one {U} the hand wants; the hand's colour need only
		// breaks ties between equal amounts, which is every ordinary
		// pick (amount one).
		n := 1
		if ch != nil {
			if v, ok := ch.ColorAmounts[cp.Color]; ok {
				n = v
			}
		}
		need := colorSymbols(st.seatHand())
		return float64(n) + 0.1*float64(need[cp.Color]), "add {" + cp.Color + "}"

	case choiceColor:
		// #742 "choose a color", #780 "…for what?". Every one of the
		// five is legal (CR 105.4), so the answer turns entirely on
		// what the card does with it — and the card says, on the
		// prompt. One policy, one switch: colorChoiceValue below.
		purpose := ""
		if ch != nil {
			purpose = ch.ColorPurpose
		}
		return st.colorChoiceValue(p.cfg, purpose, cp.Color)

	case choiceChooseSource:
		// ADR 0107 §6 decision 4: shield against the source most likely
		// to deal the damage — one an opponent controls, a spell on the
		// stack first (its damage is coming now), then the permanent
		// with the most power.
		var v float64
		for _, id := range cp.CardIDs {
			v += st.damageSourceThreat(id, lookup(id))
		}
		return v, "shield against the likeliest source"

	case choiceRingBearer:
		// ADR 0114 §7: the creature that attacks hardest carries the
		// Ring (ring_bearer.go).
		var v float64
		for _, id := range cp.CardIDs {
			v += st.ringBearerValue(lookup(id))
		}
		return v, "the Ring to the hardest hitter"

	case choicePickTarget:
		targets := cp.Targets
		if cp.Target != nil {
			targets = append(append([]targetRef(nil), targets...), *cp.Target)
		}
		return st.targetsValue(p.cfg, targets), "pick target"

	case choiceSacrifice:
		var v float64
		for _, id := range cp.CardIDs {
			if c := lookup(id); c != nil {
				v -= st.permanentValue(c)
			}
		}
		return v, "sacrifice the least"

	case choiceSearchLibrary:
		var v float64
		for _, id := range cp.CardIDs {
			v += st.cardValue(p.cfg, lookup(id))
		}
		return v, "search: take the best"

	case choiceDiscardFromHand, choiceRevealedHandPick:
		// #2115: the variant pick (exile instead, "you may choose", a
		// graveyard card) is priced the same way. Choosing nothing is
		// worth zero, so the bot takes an opponent's best card and
		// declines to give up its own when it may.
		sign := 1.0
		reason := "take their best"
		if ch == nil || ch.FromPlayer == st.me {
			// The bot is discarding its own cards.
			sign, reason = -1.0, "pitch our worst"
		}
		var v float64
		for _, id := range cp.CardIDs {
			c := lookup(id)
			v += sign * st.cardValue(p.cfg, c)
			if sign < 0 {
				// A discard the bot's own payoffs pay for
				// (discard_payoff.go).
				v += st.discardPayoff(p.cfg, c)
			}
		}
		return v, reason

	case choiceScry:
		// Bottom what the bot does not want. Everything not named in
		// Bottom stays on top in the order offered.
		var v float64
		for _, id := range cp.Bottom {
			v += p.cfg.ScryKeep - st.cardValue(p.cfg, lookup(id))
		}
		return v, "scry"

	case choicePutInLibrary:
		// ADR 0088. Two things about an ordered placement matter to
		// the bot, and the order of a pile under a library is not
		// one of them:
		//   - the card left on top of its OWN library is its next
		//     draw, so the best of them should be first; and
		//   - on "top or bottom", what it buries: its own cards the
		//     way a scry does (ScryKeep is the indifference point),
		//     and every card an opponent owns, which is how a Hinder
		//     or an Aetherspouts is meant to be played.
		//
		// #1298's shapes need no branch of their own. An exact top count
		// (Cream of the Crop): every offered answer holds the same
		// number on top, so the two terms together rank them by the card
		// left there. A look at an opponent's library (Jace's +2) and a
		// counter (Hinder) are opponents' cards, which the bury term
		// already sends down.
		var v float64
		if len(cp.TopOrder) > 0 {
			if c := lookup(cp.TopOrder[0]); c != nil && c.Owner == st.me {
				v += 0.1 * st.cardValue(p.cfg, c)
			}
		}
		for _, id := range cp.Bottom {
			if ch == nil || ch.Placement != "top_or_bottom" {
				break
			}
			c := lookup(id)
			if c != nil && c.Owner != "" && c.Owner != st.me {
				v += st.cardValue(p.cfg, c)
				continue
			}
			v += p.cfg.ScryKeep - st.cardValue(p.cfg, c)
		}
		return v, "order the library"

	case choiceChooseCards:
		// The chained-choice card-set pick, and the one kind whose
		// candidates say what naming them costs.
		//
		// #798: when the candidates are cards in the bot's OWN hand,
		// naming one is giving it up. That is every effect discard
		// since #797 — Mind Rot's forced two, a loot's discard after
		// the draw, a rummage's discard before it — and it is Sylvan
		// Library's "name two of the cards you drew" as well, which
		// puts the named cards back unless the bot pays for them. So
		// score such an answer by what it KEEPS: the total cardValue
		// of the candidates it does not name. That is the same
		// valuation the cleanup-step discard, the scry and the search
		// already use, so "the worst card in hand" means one thing in
		// this package rather than two. It settles the count for free,
		// too — keeping a card is never worth less than nothing, so an
		// "up to two" prompt takes the smallest legal answer, and a
		// loot, whose count is fixed, has no count left to settle.
		//
		// #1831: when the candidates are cards in the bot's OWN
		// library, naming one is TAKING it — the opposite sign. See
		// valueTakenFromLibrary for the callers that make that true.
		//
		// Every other choose_cards keeps the flat score. Which cards a
		// bot WANTS to name there depends entirely on what the card
		// then does with them — a Ward sacrifice, a graveyard pick, a
		// reveal — and the wire carries nothing that would say which,
		// so the enumerator's order decides. What matters is that an
		// answer is always chosen: a seat owing a choice is offered
		// nothing else, and a policy with no opinion must still pick
		// (#544).
		if v, ok := st.valueKeptInHand(p.cfg, ch, cp.CardIDs); ok {
			return v, "name the worst, keep the rest"
		}
		if v, ok := st.valueTakenFromLibrary(p.cfg, ch, cp.CardIDs); ok {
			return v, "take the best from the library"
		}
		if v, ok := st.valueTakenFromGraveyard(p.cfg, ch, cp.CardIDs); ok {
			return v, "take the best from the graveyard"
		}
		return 0.5, "choose cards"

	case choiceUntapChoice:
		// #826, CR 502.3: "choose which of these untap." The one
		// card-set pick whose sign is unambiguous — a permanent the
		// bot names UNTAPS, so the answer is worth what it untaps.
		// Scored with permanentValue, the same valuation the sacrifice
		// branch uses with the opposite sign, so "the best permanent"
		// means one thing in this package rather than two.
		//
		// That settles the count for free as well: untapping is never
		// worth less than nothing, so under a Winter Orb cap the bot
		// takes a full legal set rather than a short one, and among
		// full sets it takes the most valuable. The enumerator has
		// already filtered to sets the engine accepts (the cap solver
		// is ChooseCardsPickLegalLocked), so this is a preference over
		// legal answers and never a filter.
		var v float64
		for _, id := range cp.CardIDs {
			if c := lookup(id); c != nil {
				v += st.permanentValue(c)
			}
		}
		return v, "untap the best"

	case choiceEntryRevealFromHand:
		// #1198, CR 614.1c: "as this land enters, you may reveal an
		// Island or Swamp card from your hand. If you don't, it
		// enters tapped."
		//
		// The one card-set pick whose answer is FREE. A revealed card
		// is not a spent card — nothing moves (CR 701.20b), the card
		// is still in hand when the land has finished entering — so
		// neither of the two valuations this package already has is
		// the right one: the choose_cards branch scores an answer by
		// what it KEEPS, because naming a card there gives it up, and
		// #1028's fuel pricer prices a card eaten by a cost. Through
		// either of those, "reveal nothing" wins and every reveal-land
		// in the deck enters tapped forever.
		//
		// So: naming any card is strictly better than naming none,
		// and the enumerator has already filtered to sets the engine
		// accepts. The count settles itself — one card is enough for
		// every printed member of the family, and the enumerator
		// offers the empty answer first, so a tie can never leave the
		// land tapped.
		//
		// The information given up is real and is deliberately not
		// priced: the table learns one card in this seat's hand.
		// Against an untapped land on curve that is the trade every
		// human takes, and pricing it would need an opponent model
		// this policy does not have (docs/bot.md).
		if len(cp.CardIDs) > 0 {
			return 1, "reveal, so it enters untapped"
		}
		return 0.5, "reveal nothing"

	case choiceEntryDiscard:
		// ADR 0098, Mox Diamond: "you may discard a land card instead.
		// If you don't, put it into its owner's graveyard."
		//
		// The reveal's shape with the OPPOSITE sign on what is named:
		// a discarded card is spent. So among discards the seat names
		// its cheapest land (fuelValue, #1028 — the price of a card a
		// cost eats), and it discards rather than lose the permanent
		// unless the land is one it cannot spare: the only land card in
		// hand while it controls fewer than three lands. Declining is
		// the enumerator's AlwaysLegal answer, so either way the seat
		// has one.
		if len(cp.CardIDs) == 0 {
			return 0.5, "discard nothing: keep the land"
		}
		if st.onlyLandInHandAndShort(cp.CardIDs) {
			return 0.25, "discard: the only land in hand, and lands are short"
		}
		var fuel float64
		for _, id := range cp.CardIDs {
			fuel += p.fuelValue(st, id)
		}
		return 2 - fuel/(1+fuel), "discard the cheapest land"

	case choiceEntrySacrifice:
		if ch != nil && ch.Devour > 0 {
			// #2419: devour is a "may", any number, and each creature
			// sacrificed buys counters — see devour.go.
			return p.devourValue(st, ch, cp.CardIDs)
		}
		// ADR 0098 Decision 11, Heart of Yavimaya and Lotus Vale:
		// "sacrifice <N> instead" is not a "may", so every offered set
		// has the same size and the only question is which. Spend the
		// cheapest permanents, priced as the rest of the evaluation
		// prices them (fuelValue → permanentValue).
		var fuel float64
		for _, id := range cp.CardIDs {
			fuel += p.fuelValue(st, id)
		}
		return -fuel, "sacrifice the cheapest"

	case choiceConfirm:
		// The chained-choice two-way prompt. Both branches are always
		// legal — a confirm is a choice between two consequences, not
		// a payment the engine can refuse — so this is a preference
		// rather than a filter: take the accept branch, which is the
		// one the card's offer is for. A confirm whose accept branch
		// costs life is priced no better than any other life cost the
		// heuristic sees today (#507's flat-ActivateBase note), but
		// either answer clears the prompt, so the seat cannot stall.
		if cp.Apply != nil && *cp.Apply {
			if m.Cost != nil {
				// Same pricing the priority window uses (#547), reached
				// through a prompt instead of an activated ability:
				// Sylvan Library's "pay 4 life to keep it" is the same
				// question Griselbrand asks, and a bot at 4 life must
				// answer it the same way.
				c, refuse := p.costValue(st, nil, *m.Cost)
				if refuse {
					return suicideValue, "confirm: would pay its last life"
				}
				return 1 + c, "confirm: take the offer"
			}
			return 1, "confirm: take the offer"
		}
		return 0.5, "confirm: decline"

	case choiceCoinCall:
		// Calls have no strategic content: the engine draws won/lost rather
		// than a face, so heads and tails have identical odds even across an
		// undo. Use the pending choice's crypto-random UUID bit, the same
		// answer Layer A uses. Stop is strategic and wins only once the card
		// declared a useful-win ceiling.
		if cp.Call == "stop" {
			if ch != nil && ch.AllowStop && ch.MaxUsefulWins > 0 && ch.Wins >= ch.MaxUsefulWins {
				return 3, "coin: stop at useful wins"
			}
			return 0, "coin: keep flipping"
		}
		if ch != nil && cp.Call == aiseat.CoinCall(ch.ID) {
			return 2, "coin: random call"
		}
		return 0, "coin: other call"

	case choiceOptionalReplacement:
		// #2390: every "may" replacement, audited in replacement.go —
		// dredge, unleash and CR 903.9b have rules of their own.
		return p.optionalReplacementValue(st, ch, cp.Apply)

	case choiceEntryPayLife:
		// #2390: a shockland pays only for mana it will spend.
		return p.entryPayLifeValue(st, m, cp.Apply)

	case choiceCopyTarget:
		// #2390: "enter as a copy of" — the best permanent on offer.
		return st.copyTargetValue(cp.CardIDs, lookup)

	case choiceTriggerPrompt, choicePayUnless:
		// ADR 0104 (owner decision 8): a "yes" that TRADES the source
		// for a spell — Perplexing Chimera — is taken only when the
		// spell is worth the creature: mana value 5 or more, or a
		// permanent spell, which the bot keeps as a permanent.
		if kind == choiceTriggerPrompt && ch.TradeFor != "" {
			if !st.tradeIsWorthIt(ch.TradeFor) {
				if cp.Apply != nil && *cp.Apply {
					return 0.25, "trade: the spell is not worth the creature"
				}
				return 1, "trade: keep the creature"
			}
		}
		// The enumerator only offers "pay" when the cost is payable,
		// and a trigger the bot controls is a trigger it wants. Say
		// yes, but not so emphatically that a targeted alternative
		// in the same window cannot outbid it.
		if cp.Apply != nil && *cp.Apply {
			// ADR 0108 §5 decision 6: a payment that discards or
			// sacrifices is still paid, giving up the least valuable
			// cards. The tie-break stays above the decline's 0.5.
			if kind == choicePayUnless && ch != nil && ch.PayCards != nil && len(cp.CardIDs) > 0 {
				var loss float64
				for _, id := range cp.CardIDs {
					c := lookup(id)
					switch {
					case c == nil:
					case ch.PayCards.Action == "sacrifice":
						loss += st.permanentValue(c)
					default:
						loss += st.cardValue(p.cfg, c)
					}
				}
				if loss < 0 {
					loss = 0
				}
				return 1 - 0.4*loss/(1+loss), "pay with the least valuable"
			}
			return 1, "yes"
		}
		return 0.5, "no"

	case choiceMayCast:
		// Cascade's, discover's, suspend's, rebound's and madness's "you
		// may cast it without paying its mana cost". Taking the offer
		// only opens a free cast; whether to CAST is the ordinary
		// priority policy's call on the next decision, where the card is
		// scored like any other castable card. Since ADR 0099 a discover
		// or cascade grant closes on the seat's next pass and the card
		// goes where declining would have sent it, and a rebound grant
		// (#1854) does the same and leaves the card in exile, so
		// accepting is never worse than declining — say yes.
		if cp.Apply != nil && *cp.Apply {
			return 1, "may cast: take the free cast"
		}
		return 0.5, "may cast: decline"

	case choiceCommanderReturn:
		// ADR 0115 owner decision 2: CR 903.9a's "may" — send the
		// commander home, unless its owner could cast it from the
		// graveyard or exile it went to (escape, flashback, an
		// adventurer on an adventure). Home it is cast for its tax;
		// where it is, it may be cast without one.
		yes := cp.Apply != nil && *cp.Apply
		if ch != nil && ch.PlayableFromZone {
			if yes {
				return 0.5, "commander: castable where it is, so leave it there"
			}
			return 1, "commander: leave it where it can be cast"
		}
		if yes {
			return 1, "commander: send it to the command zone"
		}
		return 0.5, "commander: leave it"

	case choiceReplacementOrder, choiceTriggerOrder:
		// Either canonical order is as good as the other at this
		// level; take the declared one.
		return 0.5, "declared order"
	}
	return 0, "unrecognised choice"
}

// tradeIsWorthIt is owner decision 8 of ADR 0104: trade a creature for
// a spell (Perplexing Chimera) when the spell's mana value is 5 or
// more, or when it is a permanent spell — one that stays with the bot
// as a permanent. A spell the view cannot find is not worth it: the
// trade gives up a certain creature for an unknown.
func (st *state) tradeIsWorthIt(spellID string) bool {
	c := st.stack[spellID]
	if c == nil {
		return false
	}
	if manaValue(c.ManaCost, 0) >= 5 {
		return true
	}
	return !isType(c, "instant") && !isType(c, "sorcery")
}

// valueKeptInHand scores one choose_cards answer over the bot's own
// hand by the total cardValue of the candidates the answer does NOT
// name — what the bot is left holding if it answers this way. The
// second return is false for a prompt this rule has no opinion about:
// candidates on the battlefield or in a library (valueTakenFromLibrary
// prices the bot's own), or in a hand that is not the bot's, where
// naming a card is not giving it up.
//
// Cost is one cardValue per candidate per answer, and cardValue is a
// type-line switch over a card already in memory. The widest prompt
// the engine queues is a seven-card hand with "choose two" — 21
// answers, 147 of those calls — so this is linear in the answers the
// enumerator already built, with nothing of its own that grows faster.
func (st *state) valueKeptInHand(cfg Config, ch *protocol.PendingChoiceView, named []string) (float64, bool) {
	if ch == nil || len(ch.Options) == 0 || ch.FromPlayer != st.me || st.seat == nil {
		return 0, false
	}
	hand := st.seatHand()
	held := make(map[string]*protocol.CardView, len(hand))
	for i := range hand {
		held[hand[i].InstanceID] = &hand[i]
	}
	var kept float64
	for i := range ch.Options {
		id := ch.Options[i].InstanceID
		c := held[id]
		if c == nil {
			// A candidate the bot is not holding: whatever this prompt
			// is asking, it is not asking which cards to give up.
			return 0, false
		}
		if slices.Contains(named, id) {
			// Named is discarded: what the bot's own discard payoffs
			// pay for it (discard_payoff.go) is kept too.
			kept += st.discardPayoff(cfg, c)
			continue
		}
		kept += st.cardValue(cfg, c)
	}
	return kept, true
}

// libraryTakeFloor is what naming one card out of a library look is
// worth before its cardValue. A taken card is never worth less than
// nothing — it is a card the bot did not have — but cardValue prices a
// zero-cost spell, or a card the view shows redacted, at exactly zero,
// and a zero would tie with "take nothing", which the enumerator
// offers first. Small enough never to outbid a real difference
// between two cards.
const libraryTakeFloor = 0.01

// valueTakenFromLibrary scores one choose_cards answer over the bot's
// own library by the total cardValue of the candidates it NAMES (#1831)
// — the opposite sign of valueKeptInHand, because a card named out of a
// library look is a card the bot gets. The second return is false for a
// prompt this rule has no opinion about: any candidate that is not in
// the bot's own library.
//
// The zone is the whole signal, and that is a claim about the catalog
// rather than about the wire, so here is the evidence. Every
// choose_cards prompt whose candidates are cards in the chooser's own
// library comes from one of three helpers in cards/effects, and in all
// three a named card goes somewhere the chooser wants it:
//
//   - TakeFromLibraryToHand / EachPlayerTakesFromLibrary — into the
//     hand (Horn of the Mark, Explore the Vastlands);
//   - PutFromLibraryOntoBattlefield — onto the battlefield under the
//     chooser's control;
//   - hideaway — exiled face down, to be played later for free.
//
// The cards a look does NOT name go to the bottom, or stay where they
// were; nothing names a library card to mill it, exile it for good or
// bury it — those are put_in_library and scry, kinds of their own. A
// pile split over a revealed library is a reveal_pick since #1214, not
// a choose_cards. If a card ever asks "choose cards from your library"
// to put them somewhere bad, it needs a purpose on the prompt (the way
// choose_color carries ColorPurpose), and this branch must read it.
//
// Max, Min and any set rule (Explore the Vastlands' land-and/or-spell
// slots) are the enumerator's to enforce: every answer reaching this
// function is one the engine accepts, so this is a preference over
// legal sets and never a filter. "Any number" therefore takes every
// card, and "up to one" takes the best one.
func (st *state) valueTakenFromLibrary(cfg Config, ch *protocol.PendingChoiceView, named []string) (float64, bool) {
	if ch == nil || len(ch.Options) == 0 || st.seat == nil {
		return 0, false
	}
	lib := st.seat.Library.Cards
	inLibrary := make(map[string]bool, len(lib))
	for i := range lib {
		inLibrary[lib[i].InstanceID] = true
	}
	opts := make(map[string]*protocol.CardView, len(ch.Options))
	for i := range ch.Options {
		id := ch.Options[i].InstanceID
		if !inLibrary[id] {
			return 0, false
		}
		opts[id] = &ch.Options[i]
	}
	var v float64
	for _, id := range named {
		v += libraryTakeFloor + st.cardValue(cfg, opts[id])
	}
	return v, true
}

// valueTakenFromGraveyard scores one choose_cards answer over the bot's
// own graveyard by the total cardValue of the candidates it NAMES, the
// same sign as valueTakenFromLibrary (#2523, #2524). Every choose_cards
// prompt in the catalog whose candidates sit in the chooser's own
// graveyard returns or recovers the named cards — Colossal Grave-Reaver
// and Eerie Ultimatum onto the battlefield, Bound // Determined and the
// dredge family to hand, Stillness in Motion back to the library top —
// so a named card is a card the bot gets. Without this the prompt
// scored a flat 0.5 and the enumerator's first answer, "choose
// nothing", won an "any number" pick.
//
// False for any prompt with a candidate outside the bot's own
// graveyard. Min, Max and the set rule (Eerie Ultimatum's different
// names) are the enumerator's: every answer reaching here is one the
// engine accepts, so "any number" takes the largest legal set.
func (st *state) valueTakenFromGraveyard(cfg Config, ch *protocol.PendingChoiceView, named []string) (float64, bool) {
	if ch == nil || len(ch.Options) == 0 || st.seat == nil {
		return 0, false
	}
	yard := st.seat.Graveyard.Cards
	inYard := make(map[string]bool, len(yard))
	for i := range yard {
		inYard[yard[i].InstanceID] = true
	}
	opts := make(map[string]*protocol.CardView, len(ch.Options))
	for i := range ch.Options {
		id := ch.Options[i].InstanceID
		if !inYard[id] {
			return 0, false
		}
		opts[id] = &ch.Options[i]
	}
	var v float64
	for _, id := range named {
		v += libraryTakeFloor + st.cardValue(cfg, opts[id])
	}
	return v, true
}

// --- "choose a color" (#780) -----------------------------------------

// Colour purposes, as they appear on the wire (game.ColorPurpose). The
// empty string is a prompt that declares nothing, and it is handled by
// the default arm rather than by an arm of its own.
const (
	colorForMana       = "mana"
	colorForBenefit    = "benefit"
	colorForHarm       = "harm"
	colorForFilter     = "filter"
	colorForProtection = "protect"
)

// threatAttacking multiplies a creature that is already declared as an
// attacker when the protection policy ranks threats. Protection is a
// defensive answer bought a beat before it is needed, so what is coming
// at you now outranks what is merely large.
const threatAttacking = 2.0

// boardTieBreak is the ceiling of the own-board term in the benefit
// arm. It is deliberately under 0.1 — one mana symbol in hand — so the
// board only ever breaks a tie between colours the hand wants equally.
const boardTieBreak = 0.09

// colorChoiceValue is the ONE "choose a color" policy (#780), switching
// on the purpose the card declared when it asked (CR 105.4 makes all
// five colours legal, so nothing else on the prompt can say which one
// the effect wants).
//
// The UNDECLARED prompt keeps the pre-#780 rule exactly — the colour
// the bot's hand asks for most, with ties left to the enumerator's
// order, which already ranks by the colours the bot has on the
// battlefield. That is what stops a card nobody has annotated from
// regressing quietly; the catalog guard in `cards/effects` is what
// stops one from staying unannotated.
func (st *state) colorChoiceValue(cfg Config, purpose, color string) (float64, string) {
	switch purpose {
	case colorForHarm:
		// Wash Out: everything of this colour is punished, the bot's
		// own board included. Take the colour that costs the opposition
		// most net of what it costs here — which is how a mono-green
		// bot stops naming green and bouncing itself.
		return st.colorSwing(color), "harm " + color

	case colorForFilter:
		// Oona: the colour picks which of somebody else's cards the
		// effect acts on, and nothing of the bot's is at stake. The
		// best public proxy for "what is in their deck" is what they
		// have already shown — board, graveyard and stack. With nothing
		// shown every colour scores zero and the enumerator's order
		// decides, which is the fixed fallback.
		return st.opponentColorCount(color), "filter " + color

	case colorForProtection:
		// Mother of Runes, Story Circle: the colour is what is being
		// defended AGAINST, so it is the colour of the biggest thing
		// pointed this way.
		return st.biggestThreatOfColor(cfg, color), "protect from " + color

	case colorForBenefit:
		// Heraldic Banner's anthem, and Selective Obliteration's "each
		// permanent survives only if it is its controller's chosen
		// colour" — from the chooser's seat both are "name the colour
		// you want to keep". The hand still leads, because a colour the
		// bot cannot cast is a colour it will not have; the board
		// breaks the tie, which is what makes Selective Obliteration a
		// policy rather than a coincidence.
		return st.colorWanted(color) + st.ownBoardShare(color), "benefit " + color
	}
	// colorForMana, and any prompt that declares nothing.
	return st.colorWanted(color), "choose " + color
}

// colorWanted is the pre-#780 rule: how much the bot's hand is asking
// for this colour.
func (st *state) colorWanted(color string) float64 {
	return 1 + 0.1*float64(colorSymbols(st.seatHand())[color])
}

// ownBoardShare is the fraction of the bot's own permanents that are
// this colour, scaled to stay strictly under one mana symbol's worth of
// hand need. A tie-break, never a decision.
func (st *state) ownBoardShare(color string) float64 {
	var mine, total float64
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller != st.me || len(c.Colors) == 0 {
			continue
		}
		total++
		if hasColor(c, color) {
			mine++
		}
	}
	if total == 0 {
		return 0
	}
	return boardTieBreak * mine / total
}

// colorSwing is what a "hurt everything of this colour" effect buys:
// the value of the opposition's permanents of that colour, less the
// value of the bot's own. Priced with the same permanentValue the board
// evaluation uses, so an opponent's 8/8 is not traded for two of the
// bot's Islands.
func (st *state) colorSwing(color string) float64 {
	var swing float64
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if !hasColor(c, color) {
			continue
		}
		v := st.permanentValue(c)
		if c.Controller == st.me {
			swing -= v
			continue
		}
		swing += v
	}
	return swing
}

// opponentColorCount counts the opponents' cards of this colour that
// the bot is entitled to see: their battlefield, their graveyards and
// the stack. Libraries and hands are counts on the wire and stay that
// way — ADR 0033 §3 is the reason this is a proxy rather than a lookup.
func (st *state) opponentColorCount(color string) float64 {
	var n float64
	count := func(cards []protocol.CardView) {
		for i := range cards {
			c := &cards[i]
			// A graveyard card carries no controller, so fall back to
			// the owner there; on the battlefield and the stack the
			// controller is the one that matters.
			if c.Controller == st.me || (c.Controller == "" && c.Owner == st.me) {
				continue
			}
			if hasColor(c, color) {
				n++
			}
		}
	}
	count(st.view.Battlefield.Cards)
	count(st.view.Stack.Cards)
	for i := range st.view.Seats {
		if st.view.Seats[i].ID == st.me {
			continue
		}
		count(st.view.Seats[i].Graveyard.Cards)
	}
	return n
}

// biggestThreatOfColor is the value of the largest thing of this colour
// that an opponent has pointed at the table: a declared attacker first,
// then any creature, then any other permanent, then a spell on the
// stack. Zero for a colour nobody is threatening in, so a quiet board
// leaves the enumerator's order to decide.
func (st *state) biggestThreatOfColor(cfg Config, color string) float64 {
	var best float64
	consider := func(v float64) {
		if v > best {
			best = v
		}
	}
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller == st.me || !hasColor(c, color) {
			continue
		}
		if !isCreature(c) {
			consider(st.permanentValue(c))
			continue
		}
		v := st.w.CombatValue(c)
		if c.AttackingTarget != "" {
			v *= threatAttacking
		}
		consider(v)
	}
	for i := range st.view.Stack.Cards {
		c := &st.view.Stack.Cards[i]
		if c.Controller == st.me || !hasColor(c, color) {
			continue
		}
		consider(st.cardValue(cfg, c))
	}
	return best
}

// seatHand is the bot's own hand, or nil. Used for colour-preference
// counting; an empty hand simply yields no preference.
func (st *state) seatHand() []protocol.CardView {
	if st.seat == nil {
		return nil
	}
	return st.seat.Hand.Cards
}

// entryControllerValue scores one seat of an entry_controller prompt —
// "this enters under the control of an opponent of your choice"
// (ADR 0102). The prompt's control_purpose says which way round the
// gift cuts (owner decision 6, 2026-09-30):
//
//   - "harm" (Captive Audience, Xantcha): the STRONGEST opponent, by
//     SeatEval.Strength. Not Threat, which ranks a seat by how close it
//     is to dying — "your life total becomes 4" costs a seat at 5 life
//     almost nothing, and punishes the leader most.
//   - "benefit" (Pendant of Prosperity): the WEAKEST opponent, so the
//     help goes where it threatens the bot least.
//
// A seat the view has no evaluation for scores as neither, so the
// enumerator's first offered seat wins the tie.
func (st *state) entryControllerValue(ch *protocol.PendingChoiceView, index *int) (float64, string) {
	if ch == nil || index == nil || *index < 0 || *index >= len(ch.PickOptions) {
		return 0, "entry controller: first offered opponent"
	}
	opt := ch.PickOptions[*index]
	e := st.evals[opt.Player]
	if e == nil {
		return 0, "entry controller: " + opt.Label
	}
	if ch.ControlPurpose == "benefit" {
		return -e.Strength, "give it to the weakest opponent (" + opt.Label + ")"
	}
	return e.Strength, "give it to the strongest opponent (" + opt.Label + ")"
}
