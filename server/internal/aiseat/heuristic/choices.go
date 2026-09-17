package heuristic

import (
	"context"

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
	choiceColor               = "choose_color"
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
		// The chained-choice card-set pick. Which cards a bot WANTS to
		// name depends entirely on what the card then does with them —
		// Sylvan Library's picks are the cards it pays life for, so
		// neither "take the best" nor "give up the worst" is right as
		// a general rule, and the wire carries nothing that would say
		// which. Score every offered set the same and let the
		// enumerator's order decide. What matters here is that an
		// answer is always chosen: a seat owing a choice is offered
		// nothing else, and a policy with no opinion must still pick
		// (#544).
		return 0.5, "choose cards"

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

// seatHand is the bot's own hand, or nil. Used for colour-preference
// counting; an empty hand simply yields no preference.
func (st *state) seatHand() []protocol.CardView {
	if st.seat == nil {
		return nil
	}
	return st.seat.Hand.Cards
}
