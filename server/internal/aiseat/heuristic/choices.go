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
	choicePutInLibrary        = "put_in_library"
	choiceSearchLibrary       = "search_library"
	choiceEntryPayLife        = "entry_pay_life"
	choiceConfirm             = "confirm"
	choiceChooseCards         = "choose_cards"
	choiceUntapChoice         = "untap_choice"
	choiceEntryRevealFromHand = "entry_reveal_from_hand"
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
		// #742 "choose a color", #780 "…for what?". Every one of the
		// five is legal (CR 105.4), so the answer turns entirely on
		// what the card does with it — and the card says, on the
		// prompt. One policy, one switch: colorChoiceValue below.
		purpose := ""
		if ch != nil {
			purpose = ch.ColorPurpose
		}
		return st.colorChoiceValue(p.cfg, purpose, cp.Color)

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
