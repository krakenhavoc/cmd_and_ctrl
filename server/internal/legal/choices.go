package legal

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// choiceParams is the resolve_choice wire payload. The dispatcher
// routes on which optional field is present, so each kind below
// fills exactly the field(s) its resolver reads.
type choiceParams struct {
	ChoiceID    string        `json:"choice_id"`
	CardIDs     []string      `json:"card_ids,omitempty"`
	Color       string        `json:"color,omitempty"`
	Order       []string      `json:"order,omitempty"`
	Apply       *bool         `json:"apply,omitempty"`
	Assignments []assignParam `json:"assignments,omitempty"`
	TrampleTo   int           `json:"trample_to_player,omitempty"`
	Target      *targetWire   `json:"target,omitempty"`
	Targets     []targetWire  `json:"targets"`
	Bottom      []string      `json:"bottom"`
	TopOrder    []string      `json:"top_order"`
	// Graveyard is the surveil answer's bin leg. No omitempty, for
	// the same reason Bottom/TopOrder have none: "keep all on top"
	// is the empty slice, and omitempty would erase it into absent,
	// which is how the dispatcher tells a surveil from a scry.
	Graveyard []string `json:"graveyard"`
}

type assignParam struct {
	BlockerID string `json:"blocker_id"`
	Amount    int    `json:"amount"`
}

// choiceMoves enumerates answers to every pending choice owed by
// this seat. Returns true when at least one choice is owed — the
// caller then offers nothing else, because the engine refuses
// pass_priority while any choice is open.
//
// Ordering choices (replacement_order, trigger_order) and scry are
// permutation-shaped; they are offered as a small canonical set
// rather than every permutation: the declared order, its reverse,
// and for scry each "bottom exactly this card" plus keep-all and
// bottom-all. Damage assignment is offered as the single canonical
// lethal-in-order split (plus trample overflow), which is what a
// player does almost every time.
func (e *enumerator) choiceMoves() bool {
	g := e.g
	owed := false
	for _, c := range g.PendingChoices {
		if c == nil || c.Chooser != e.seat {
			continue
		}
		owed = true
		base := func() choiceParams { return choiceParams{ChoiceID: c.ID.String()} }
		reason := c.Reason
		if reason == "" {
			reason = string(c.Kind)
		}
		switch c.Kind {
		case game.PendingChoiceDiscardFromHand:
			from := playerByID(g, c.FromPlayer)
			if from == nil || from.Hand == nil {
				continue
			}
			var pool []uuid.UUID
			for _, h := range from.Hand.Cards {
				pool = append(pool, h.InstanceID)
			}
			for _, set := range combinations(pool, c.Count, c.Count, e.opts.MaxExpansionPerSource) {
				p := base()
				p.CardIDs = idStrings(set)
				label := reason + ": discard"
				for _, id := range set {
					label += " " + cardName(g, id)
				}
				e.addChoice(c, label, p)
			}
			if c.Count == 0 {
				p := base()
				p.CardIDs = []string{}
				e.addChoice(c, reason+": discard nothing", p)
			}

		case game.PendingChoiceMana:
			for _, color := range c.ColorOptions {
				p := base()
				p.Color = color
				e.addChoice(c, reason+": add {"+color+"}", p)
			}

		case game.PendingChoiceReplacementOrder:
			ids := c.ReplacementEffectIDs
			for _, order := range canonicalOrders(len(ids)) {
				p := base()
				for _, i := range order {
					p.Order = append(p.Order, strconv.FormatUint(uint64(ids[i]), 10))
				}
				e.addChoice(c, fmt.Sprintf("%s: apply replacements in order %v", reason, order), p)
			}

		case game.PendingChoiceTriggerOrder:
			ids := c.TriggerOrderIDs
			for _, order := range canonicalOrders(len(ids)) {
				p := base()
				for _, i := range order {
					p.Order = append(p.Order, ids[i].String())
				}
				e.addChoice(c, fmt.Sprintf("%s: resolve triggers in order %v", reason, order), p)
			}

		case game.PendingChoiceOptionalReplacement, game.PendingChoiceTriggerPrompt:
			for _, apply := range []bool{true, false} {
				a := apply
				p := base()
				p.Apply = &a
				verb := "yes"
				if !apply {
					verb = "no"
				}
				e.addChoice(c, reason+": "+verb, p)
			}

		case game.PendingChoicePayUnless:
			// Paying is only a move if the cost is payable; the engine
			// would silently treat an unpayable "yes" as a decline.
			canPay := false
			if cost, err := game.ParseCost(c.PayCost); err == nil {
				// Zero spend context, matching payCostLocked: a
				// pay-unless cost is neither a cast nor an
				// activation, so restricted mana cannot fund it
				// (#352).
				canPay = e.canPay(cost, 0, game.ManaSpendContext{})
			}
			for _, apply := range []bool{true, false} {
				if apply && !canPay {
					continue
				}
				a := apply
				p := base()
				p.Apply = &a
				verb := "decline"
				if apply {
					verb = "pay " + c.PayCost
				}
				e.addChoice(c, reason+": "+verb, p)
			}

		case game.PendingChoiceDamageAssignment:
			if m, ok := e.canonicalDamageAssignment(c, base()); ok {
				e.addChoice(c, reason+": assign combat damage", m)
			}

		case game.PendingChoicePickTarget:
			cands := make([]game.TargetRef, 0, len(c.PickTargetPlayers)+len(c.PickTargetCards))
			for _, id := range c.PickTargetPlayers {
				cands = append(cands, game.TargetRef{Kind: game.TargetPlayer, ID: id})
			}
			for _, id := range c.PickTargetCards {
				cands = append(cands, game.TargetRef{Kind: game.TargetCard, ID: id})
			}
			lo, hi := c.PickTargetMin, c.PickTargetMax
			if hi <= 0 || hi > len(cands) {
				hi = len(cands)
			}
			if lo == 0 && hi == 1 || lo == 1 && hi == 1 {
				// Single-slot prompt: the resolver reads `target`.
				for _, t := range cands {
					p := base()
					tw := targetWire{Kind: string(t.Kind), ID: t.ID.String()}
					p.Target = &tw
					e.addChoice(c, reason+targetLabel(g, []game.TargetRef{t}), p)
				}
				continue
			}
			for _, set := range combinationsRefs(cands, lo, hi, e.opts.MaxExpansionPerSource) {
				p := base()
				p.Targets = wireTargets(set)
				if p.Targets == nil {
					p.Targets = []targetWire{}
				}
				e.addChoice(c, reason+targetLabel(g, set), p)
			}

		case game.PendingChoiceSacrifice:
			for _, id := range c.SacrificeOptions {
				if card := findBattlefield(g, id); card == nil || card.Controller != e.seat {
					continue
				}
				p := base()
				p.CardIDs = []string{id.String()}
				e.addChoice(c, reason+": sacrifice "+cardName(g, id), p)
			}

		case game.PendingChoiceCopyTarget:
			// CR 614.1c: every printed "enters as a copy" says "you
			// may", so declining is always a legal answer — and for
			// a bot it is the one that always terminates, which is
			// why it is enumerated first.
			p := base()
			p.CardIDs = []string{}
			e.addChoice(c, reason+": don't copy anything", p)
			for _, id := range c.CopyOptions {
				if card := findBattlefield(g, id); card == nil {
					continue
				}
				p := base()
				p.CardIDs = []string{id.String()}
				e.addChoice(c, reason+": copy "+cardName(g, id), p)
			}

		case game.PendingChoiceSearchLibrary:
			// CR 701.19: take up to SearchMax of the matching cards;
			// failing to find (an empty list) is always legal.
			p := base()
			p.CardIDs = []string{}
			e.addChoice(c, reason+": fail to find", p)
			for _, set := range combinations(c.SearchCards, 1, c.SearchMax, e.opts.MaxExpansionPerSource) {
				p := base()
				p.CardIDs = idStrings(set)
				label := reason + ": take"
				for _, id := range set {
					label += " " + cardName(g, id)
				}
				e.addChoice(c, label, p)
			}

		case game.PendingChoiceEntryPayLife:
			// "As this enters, you may pay N life." The engine only
			// asks when the payer can afford it, and a pay it can no
			// longer afford degrades to a decline; both answers are
			// always accepted.
			for _, apply := range []bool{true, false} {
				a := apply
				p := base()
				p.Apply = &a
				verb := "enter tapped"
				if apply {
					verb = "pay " + c.PayCost
				}
				e.addChoice(c, reason+": "+verb, p)
			}

		case game.PendingChoiceScry:
			cards := c.ScryCards
			all := idStrings(cards)
			// Keep all, in order.
			p := base()
			p.TopOrder = all
			p.Bottom = []string{}
			e.addChoice(c, reason+": keep all on top", p)
			if len(cards) > 0 {
				// Bottom all.
				p = base()
				p.Bottom = all
				p.TopOrder = []string{}
				e.addChoice(c, reason+": bottom all", p)
			}
			if len(cards) > 1 {
				// Bottom exactly one, keep the rest in order.
				for i, id := range cards {
					p = base()
					p.Bottom = []string{id.String()}
					for j, other := range cards {
						if j != i {
							p.TopOrder = append(p.TopOrder, other.String())
						}
					}
					e.addChoice(c, reason+": bottom "+cardName(g, id), p)
				}
			}

		case game.PendingChoiceSurveil:
			// Same permutation shape as scry, with "bottom" replaced
			// by "graveyard" — and the canonical set matters more
			// here, because "bin all" and "keep all" are the two
			// answers a surveil deck actually gives.
			cards := c.ScryCards
			all := idStrings(cards)
			p := base()
			p.TopOrder = all
			p.Graveyard = []string{}
			e.addChoice(c, reason+": keep all on top", p)
			if len(cards) > 0 {
				p = base()
				p.Graveyard = all
				p.TopOrder = []string{}
				e.addChoice(c, reason+": all to graveyard", p)
			}
			if len(cards) > 1 {
				for i, id := range cards {
					p = base()
					p.Graveyard = []string{id.String()}
					for j, other := range cards {
						if j != i {
							p.TopOrder = append(p.TopOrder, other.String())
						}
					}
					e.addChoice(c, reason+": "+cardName(g, id)+" to graveyard", p)
				}
			}

		case game.PendingChoiceLookAtTop:
			// No away lane, so the answer space is permutations of
			// the looked-at cards. Offer the two that matter — leave
			// it alone, and pull each card to the front — rather
			// than N! entries.
			cards := c.ScryCards
			p := base()
			p.TopOrder = idStrings(cards)
			e.addChoice(c, reason+": leave the order alone", p)
			for i, id := range cards {
				if i == 0 {
					continue
				}
				p = base()
				p.TopOrder = []string{id.String()}
				for j, other := range cards {
					if j != i {
						p.TopOrder = append(p.TopOrder, other.String())
					}
				}
				e.addChoice(c, reason+": put "+cardName(g, id)+" on top", p)
			}
		}
	}
	return owed
}

func (e *enumerator) addChoice(c *game.PendingChoice, label string, p choiceParams) {
	e.add(Move{
		Type:   TypeResolveChoice,
		Player: e.seat,
		Kind:   KindChoice,
		Label:  label,
		Source: c.Source,
		Params: mustJSON(p),
	})
}

// canonicalOrders returns the identity permutation and, when there
// is more than one element, its reverse. Enough for a bot; a human
// keeps the drag-to-reorder UI.
func canonicalOrders(n int) [][]int {
	if n == 0 {
		return [][]int{{}}
	}
	id := make([]int, n)
	for i := range id {
		id[i] = i
	}
	if n == 1 {
		return [][]int{id}
	}
	rev := make([]int, n)
	for i := range rev {
		rev[i] = n - 1 - i
	}
	return [][]int{id, rev}
}

// canonicalDamageAssignment builds the one split the prefix-lethal
// rule always accepts: walk the blockers in declared order, give each
// lethal (1 with deathtouch, else remaining toughness) until power
// runs out, dump the remainder on the last blocker — or, with
// trample, on the defending player.
func (e *enumerator) canonicalDamageAssignment(c *game.PendingChoice, p choiceParams) (choiceParams, bool) {
	f := c.DamageAssignment
	if f == nil {
		return p, false
	}
	remaining := f.AttackerPower
	assigns := make([]assignParam, 0, len(f.BlockerIDs))
	for _, id := range f.BlockerIDs {
		lethal := 1
		if !f.HasDeathtouch {
			if b := findBattlefield(e.g, id); b != nil {
				lethal = b.CurrentToughness() - b.DamageMarked
				if lethal < 1 {
					lethal = 1
				}
			}
		}
		give := lethal
		if give > remaining {
			give = remaining
		}
		assigns = append(assigns, assignParam{BlockerID: id.String(), Amount: give})
		remaining -= give
	}
	if remaining > 0 {
		if f.AllowTrample {
			p.TrampleTo = remaining
		} else if len(assigns) > 0 {
			assigns[len(assigns)-1].Amount += remaining
		} else {
			return p, false
		}
	}
	if len(assigns) == 0 && p.TrampleTo == 0 && f.AttackerPower > 0 {
		return p, false
	}
	p.Assignments = assigns
	return p, true
}

// combinationsRefs is combinations over TargetRefs, including the
// empty set when lo is 0.
func combinationsRefs(cands []game.TargetRef, lo, hi, limit int) [][]game.TargetRef {
	var out [][]game.TargetRef
	if lo == 0 {
		out = append(out, nil)
	}
	start := lo
	if start < 1 {
		start = 1
	}
	if hi > len(cands) {
		hi = len(cands)
	}
	for k := start; k <= hi && len(out) < limit; k++ {
		var rec func(s int, cur []game.TargetRef)
		rec = func(s int, cur []game.TargetRef) {
			if len(out) >= limit {
				return
			}
			if len(cur) == k {
				out = append(out, append([]game.TargetRef(nil), cur...))
				return
			}
			for i := s; i < len(cands); i++ {
				rec(i+1, append(cur, cands[i]))
			}
		}
		rec(0, nil)
	}
	return out
}
