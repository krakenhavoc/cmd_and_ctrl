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
	choiceSearchLibrary       = "search_library"
	choiceEntryPayLife        = "entry_pay_life"
	choiceConfirm             = "confirm"
	choiceChooseCards         = "choose_cards"
	choiceUntapChoice         = "untap_choice"
	choiceColor               = "choose_color"
	choiceCoinCall            = "coin_call"
)

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
			v -= st.cardValue(p.cfg, st.mine[id])
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
		// #742 "choose a color": the colour the bot's hand asks for
		// most. A tie keeps the enumerator's order, which already
		// ranks by the colours the bot has on the battlefield — so an
		// empty hand still names the board's main colour for a
		// Coldsteel Heart rather than white by default.
		need := colorSymbols(st.seatHand())
		return 1 + 0.1*float64(need[cp.Color]), "choose " + cp.Color

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
				v -= st.w.permanentValue(c)
			}
		}
		return v, "sacrifice the least"

	case choiceSearchLibrary:
		var v float64
		for _, id := range cp.CardIDs {
			v += st.cardValue(p.cfg, lookup(id))
		}
		return v, "search: take the best"

	case choiceDiscardFromHand:
		sign := 1.0
		reason := "take their best"
		if ch == nil || ch.FromPlayer == st.me {
			// The bot is discarding its own cards.
			sign, reason = -1.0, "pitch our worst"
		}
		var v float64
		for _, id := range cp.CardIDs {
			v += sign * st.cardValue(p.cfg, lookup(id))
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
		// Every other choose_cards keeps the flat score. Which cards a
		// bot WANTS to name there depends entirely on what the card
		// then does with them — a Ward sacrifice, a library pick, a
		// reveal — and the wire carries nothing that would say which,
		// so the enumerator's order decides. What matters is that an
		// answer is always chosen: a seat owing a choice is offered
		// nothing else, and a policy with no opinion must still pick
		// (#544).
		if v, ok := st.valueKeptInHand(p.cfg, ch, cp.CardIDs); ok {
			return v, "name the worst, keep the rest"
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
				v += st.w.permanentValue(c)
			}
		}
		return v, "untap the best"

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

	case choiceTriggerPrompt, choiceOptionalReplacement, choiceEntryPayLife, choicePayUnless:
		// The enumerator only offers "pay" when the cost is payable,
		// and a trigger the bot controls is a trigger it wants. Say
		// yes, but not so emphatically that a targeted alternative
		// in the same window cannot outbid it.
		if cp.Apply != nil && *cp.Apply {
			return 1, "yes"
		}
		return 0.5, "no"

	case choiceReplacementOrder, choiceTriggerOrder:
		// Either canonical order is as good as the other at this
		// level; take the declared one.
		return 0.5, "declared order"
	}
	return 0, "unrecognised choice"
}

// valueKeptInHand scores one choose_cards answer over the bot's own
// hand by the total cardValue of the candidates the answer does NOT
// name — what the bot is left holding if it answers this way. The
// second return is false for a prompt this rule has no opinion about:
// candidates on the battlefield or in a library, or in a hand that is
// not the bot's, where naming a card is not giving it up.
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
			continue
		}
		kept += st.cardValue(cfg, c)
	}
	return kept, true
}

// seatHand is the bot's own hand, or nil. Used for colour-preference
// counting; an empty hand simply yields no preference.
func (st *state) seatHand() []protocol.CardView {
	if st.seat == nil {
		return nil
	}
	return st.seat.Hand.Cards
}
