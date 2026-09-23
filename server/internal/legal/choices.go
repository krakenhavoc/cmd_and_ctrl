package legal

import (
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"

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
	Call        string        `json:"call,omitempty"`
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
	// Iterations answers a loop_shortcut prompt (CR 726): how many
	// more times the loop's controller wants the repeating ability to
	// resolve. omitempty is safe here where it would be wrong on the
	// scry keys: the dispatcher routes this kind by the choice's KIND,
	// and the value that goes missing — zero — is the one the
	// resolver reads as "stop here" anyway.
	Iterations int `json:"iterations,omitempty"`
	// OptionIndex answers an option_pick prompt (#568): which of the
	// prompt's branches the chooser took. A POINTER, for the reason
	// Apply is one — the meaningful value is zero (the first option),
	// so an int field with omitempty would erase the commonest answer
	// and a plain int would be sent on every other kind. The
	// dispatcher still routes this kind by KIND, not by presence.
	OptionIndex *int `json:"option_index,omitempty"`
	// Modes answers a mode_pick prompt (#764, CR 603.3c): the chosen
	// OPTION indexes in the order chosen, repeats allowed when the
	// ability says so. No omitempty, and the dispatcher routes this
	// kind by the choice's KIND — "choose up to one, and I choose
	// none" is the empty slice, which omitempty would erase into
	// absent, and an index list of zeroes is a perfectly ordinary
	// answer ([0, 0, 0] is Mystic Confluence drawing three cards).
	Modes []int `json:"modes"`
	// CreatureType answers a choose_creature_type prompt (CR 614.12).
	// omitempty because the dispatcher routes on its PRESENCE: an
	// empty string sent on every other kind would be read as "this is
	// a creature-type answer".
	CreatureType string `json:"creature_type,omitempty"`
	// CardName answers a choose_card_name prompt (CR 614.12, #1210).
	// omitempty for CreatureType's reason: the dispatcher routes on
	// its PRESENCE.
	CardName string `json:"card_name,omitempty"`
}

type assignParam struct {
	BlockerID string `json:"blocker_id"`
	Amount    int    `json:"amount"`
}

// choiceMoves enumerates answers to every pending choice owed by
// this seat. Returns true when at least one choice is owed — the
// caller then offers nothing else, because the engine refuses
// pass_priority while a blocking choice is open, and because a
// question already in front of a seat is the one thing it should be
// answering. That holds for a NON-blocking prompt too (#794): a bot
// owing a Rhystic tax is offered pay / decline, not the rest of its
// turn — one dispatch clears it and the next window has everything.
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
					// cardNameFor, not cardName: the pool is
					// FromPlayer's hand, which is only this seat's to
					// read when an effect revealed it (ADR 0033 §3).
					label += " " + cardNameFor(g, id, e.seat)
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
				// #742: a one-pick-N-mana choice names every token it
				// adds, so "add {G}{G}{G}" and "add {G}" are told apart.
				n := 1
				if v, ok := c.ManaAmounts[color]; ok {
					n = v
				}
				e.addChoice(c, reason+": add "+strings.Repeat("{"+color+"}", n), p)
			}

		case game.PendingChoiceColor:
			// #742. "Choose a color" (CR 105.4), stored ("as this
			// enters") or at resolution. ResolveColorChoice accepts
			// every colour on the choice's own option list and treats
			// a departed source as "nowhere to land", so every answer
			// is always legal and the seat is never handed an empty
			// list — at least one colour is always on offer, since the
			// narrowest printed list ("other than blue") has four.
			//
			// Ordered by the prompt's own ColorPurpose (#986):
			// `harm` by what the OPPOSITION loses net of what this
			// seat does, `filter` by what the opposition has shown,
			// `protect` by the biggest threat pointed this way, and
			// `mana` / `benefit` / undeclared by what this seat
			// already has — which is what a player choosing for
			// Coldsteel Heart or Heraldic Banner does. Battlefield
			// counts only, so the ranking reads nothing hidden, and
			// every option is still offered whatever the order: a
			// policy that wants a different colour still sees all
			// five. The same function orders the human's buttons.
			for _, color := range e.colorAnswers(c) {
				p := base()
				p.Color = color
				e.addAlwaysLegalChoice(c, reason+": "+game.ColorName(color), p)
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

		// S27's two non-targeting prompts answer with exactly the
		// pick_target payload — a single {kind, id} ref out of a
		// server-computed set, min 1 max 1 — and actions.go routes
		// them apart on the choice KIND rather than on the shape.
		// They belong on this case for the same reason they reused
		// that payload, and leaving them off it was a bot hang of the
		// #544 class in the other direction: not an enumerated move
		// the engine refuses, but a prompt with NO enumerated answer,
		// which stops the seat dead because the engine also refuses
		// pass_priority while a choice is open. Any seat that cast a
		// battle, or that controlled two legends, was stuck.
		case game.PendingChoicePickTarget, game.PendingChoiceChooseProtector, game.PendingChoiceLegendRule:
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

		// #1196 CR 115.7: "choose the new target for …". One slot of
		// an item already on the stack, so the answer is one ref (or
		// none, when the prompt allows declining) and never a
		// combination — which is why it is its own case rather than
		// another kind on the pick_target one above.
		//
		// Ordered through #1014's Options.OrderTargets, and the
		// DECLINE comes first when it is legal: a bot that has to
		// choose between moving a removal spell and leaving it where
		// it is should have "leave it" in the list at all, and it is
		// also the answer that always terminates.
		case game.PendingChoiceRetarget:
			cands := make([]game.TargetRef, 0, len(c.PickTargetPlayers)+len(c.PickTargetCards))
			for _, id := range c.PickTargetPlayers {
				cands = append(cands, game.TargetRef{Kind: game.TargetPlayer, ID: id})
			}
			for _, id := range c.PickTargetCards {
				cands = append(cands, game.TargetRef{Kind: game.TargetCard, ID: id})
			}
			e.orderCandidates(cands)
			if c.PickTargetMin == 0 {
				p := base()
				p.Targets = []targetWire{}
				e.addChoice(c, reason+": leave the target unchanged", p)
			}
			for _, t := range cands {
				p := base()
				p.Targets = wireTargets([]game.TargetRef{t})
				e.addChoice(c, reason+targetLabel(g, []game.TargetRef{t}), p)
			}

		// #764 CR 603.3c: a modal trigger's mode, chosen as the
		// ability is put on the stack. Every legal selection, bounded
		// by MaxExpansionPerSource and ordered by ADR 0065 §6 — the
		// all-one-option selections first, so a repeatable prompt
		// whose only legal option is one mode is not crowded out by
		// mixed multisets it cannot take.
		case game.PendingChoiceModePick:
			for _, sel := range game.ModePickSelections(c, e.opts.MaxExpansionPerSource) {
				p := base()
				p.Modes = sel
				if p.Modes == nil {
					p.Modes = []int{}
				}
				e.addChoice(c, reason+modeLabel(c, sel), p)
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
			// CR 701.23: take up to SearchMax of the matching cards;
			// failing to find (an empty list) is always legal, which
			// is what makes it this choice's AlwaysLegal answer.
			p := base()
			p.CardIDs = []string{}
			e.addAlwaysLegalChoice(c, reason+": fail to find", p)
			// Every offered pick is run past the engine's own
			// acceptance check first. A search may carry a Validate
			// hook for a clause no per-card predicate can express —
			// Myriad Landscape's "two basic lands that SHARE a land
			// type" — and enumerating a pick the resolver will reject
			// breaks this package's one promise (#544). The rule is
			// not re-derived here; the engine is asked.
			picks := filteredCombinations(c.SearchCards, 1, c.SearchMax, e.opts.MaxExpansionPerSource,
				func(set []uuid.UUID) bool { return g.SearchPickLegalLocked(c, set) })
			for _, set := range picks {
				p := base()
				p.CardIDs = idStrings(set)
				label := reason + ": take"
				for _, id := range set {
					// Library cards: named only because the search
					// marked this seat a knower of every match.
					label += " " + cardNameFor(g, id, e.seat)
				}
				e.addChoice(c, label, p)
			}

		case game.PendingChoiceCoinCall:
			for _, call := range []string{"heads", "tails"} {
				p := base()
				p.Call = call
				if call == "heads" {
					e.addAlwaysLegalChoice(c, reason+": call "+call, p)
				} else {
					e.addChoice(c, reason+": call "+call, p)
				}
			}
			if c.CoinAllowStop {
				p := base()
				p.Call = "stop"
				e.addChoice(c, reason+": stop flipping", p)
			}

		case game.PendingChoiceConfirm:
			// The chained-choice two-way prompt. Both branches are
			// offered and ResolveConfirm validates nothing about the
			// board — a confirm is a choice between two consequences
			// the card already committed to, not a payment the engine
			// can refuse. So the decline is this kind's AlwaysLegal
			// answer: a seat whose every other answer keeps bouncing
			// has somewhere to go instead of going to sleep (#544).
			//
			// Marked on the DECLINE rather than the accept because the
			// accept branch is the one a card hangs a cost on.
			for _, apply := range []bool{true, false} {
				a := apply
				p := base()
				p.Apply = &a
				verb := c.DeclineLabel
				if apply {
					verb = c.AcceptLabel
				}
				if verb == "" {
					verb = "no"
					if apply {
						verb = "yes"
					}
				}
				if apply {
					// The accept branch carries whatever the card
					// charges for it (#547's MoveCost), so a policy
					// holding only the wire payload does not price
					// "pay 4 life to keep this" like "shuffle".
					e.add(Move{
						Type:   TypeResolveChoice,
						Player: e.seat,
						Kind:   KindChoice,
						Label:  reason + ": " + verb,
						Source: c.Source,
						Params: mustJSON(p),
						Cost:   moveCost(c.LifeCost, 0),
					})
					continue
				}
				e.addAlwaysLegalChoice(c, reason+": "+verb, p)
			}

		case game.PendingChoiceOptionPick:
			// #568. "Choose one of the following", addressed to any
			// seat — Torment of Hailfire's three-way question, and
			// the pile a Fact or Fiction chooser takes.
			//
			// ResolveOptionPick validates the INDEX and nothing else,
			// so every offered option is an answer the engine will
			// accept. The legality lives at queue time: an effect
			// builds the list out of what this seat can actually do.
			// So every option is offered, in the card's printed order,
			// and the FIRST is the always-legal way out — the kind's
			// contract is that a queuing effect puts a branch that
			// always works there ("lose 3 life" on Torment, which needs
			// no permanent and no card in hand).
			//
			// Each answer carries the branch's declared life cost as
			// MoveCost, for #547's reason: a policy holding only the
			// wire payload otherwise prices "lose 3 life" exactly like
			// "discard a card", and a bot at 3 life answers with the
			// life and dies. Cheapest-by-cost is then the policy's
			// decision, not the enumerator's. See docs/bot.md.
			for i, opt := range c.PickOptions {
				p := base()
				idx := i
				p.OptionIndex = &idx
				label := opt.Label
				if label == "" {
					label = "option " + strconv.Itoa(i+1)
				}
				if i == 0 {
					e.addAlwaysLegalChoice(c, reason+": "+label, p)
					continue
				}
				e.add(Move{
					Type:   TypeResolveChoice,
					Player: e.seat,
					Kind:   KindChoice,
					Label:  reason + ": " + label,
					Source: c.Source,
					Params: mustJSON(p),
					Cost:   moveCost(opt.LifeCost, 0),
				})
			}

		case game.PendingChoiceChooseCards, game.PendingChoiceUntapChoice,
			game.PendingChoiceEntryRevealFromHand, game.PendingChoiceRevealPick,
			game.PendingChoiceTheirPermanents, game.PendingChoiceOwnPermanents:
			// "Choose N of these cards." The bounds ride on the
			// choice, and a prompt may also carry a set-level
			// Validate hook ("discard two unless you discard a
			// creature card") that lives on the unexported
			// continuation frame. So every set is run past the
			// engine's own acceptance check before it is offered —
			// ChooseCardsPickLegalLocked, the same window search
			// gets — rather than enumerating a superset the
			// enumerator cannot see the constraint on, which is the
			// hole #544 fell into.
			//
			// The empty answer is offered first when the floor is
			// zero, for the same reason search offers "fail to find"
			// first: it is the answer that always terminates, and the
			// engine skips both the zone re-check and Validate for
			// it, so it is this kind's AlwaysLegal answer.
			//
			// A prompt with a floor above zero has NO unconditional
			// answer — every set it accepts is a set of cards that
			// must still be where the prompt found them and must pass
			// the card's rule — and is deliberately left unmarked
			// rather than given a synthesised fallback the resolver
			// would refuse.
			if c.ChooseMin == 0 {
				p := base()
				p.CardIDs = []string{}
				e.addAlwaysLegalChoice(c, reason+": choose nothing", p)
			}
			// filteredCombinations, not combinations() plus a filter:
			// it spreads the budget across set sizes, so a "choose one
			// or two" prompt whose rule rejects most single cards
			// still reaches the valid pairs instead of spending the
			// whole budget on singles (#544, and
			// TestChooseCardsReachesValidPairsPastTheBudget).
			sets := filteredCombinations(e.cardSetPickPool(c), c.ChooseMin, c.ChooseMax, e.opts.MaxExpansionPerSource,
				func(set []uuid.UUID) bool { return g.ChooseCardsPickLegalLocked(c, set) })
			// #826: untap_choice shares every line of this branch —
			// the same payload, the same bounds and the same
			// engine-side acceptance check, which is where the cap
			// solver lives (ADR 0070 Decision 3). Only the verb the
			// seat reads differs.
			//
			// #1198: and so does entry_reveal_from_hand, CR 614.1c's
			// "you may reveal an Island or Swamp card from your
			// hand". Its floor is always zero, so the empty answer
			// above is this kind's AlwaysLegal move and the prompt
			// can never leave a seat without one — which matters
			// more here than anywhere else on this branch, because
			// the seat owing it is holding a permanent halfway onto
			// the battlefield.
			//
			// #1214's three resolution-time picks join on exactly the
			// same terms, and add one thing of their own: the ORDER
			// the pool is walked in, which is what decides which sets
			// survive MaxExpansionPerSource (see cardSetPickPool).
			verb := ": choose"
			switch c.Kind {
			case game.PendingChoiceUntapChoice:
				verb = ": untap"
			case game.PendingChoiceEntryRevealFromHand:
				verb = ": reveal"
			case game.PendingChoiceTheirPermanents:
				// The label says WHOSE board is on offer, because a
				// run of these is one prompt per player and four
				// identical "choose Grizzly Bears" lines is a move
				// list a human reading a bot's log cannot follow.
				verb = ": choose from " + choiceSeatName(g, c.FromPlayer)
			}
			for _, set := range sets {
				p := base()
				p.CardIDs = idStrings(set)
				label := reason + verb
				for _, id := range set {
					// cardNameFor, not cardName: the candidates are
					// as often cards in a hand as cards on the
					// battlefield, and only a seat the effect made a
					// knower may read them (ADR 0033 §3).
					label += " " + cardNameFor(g, id, e.seat)
				}
				e.addChoice(c, label, p)
			}
			if c.ChooseMin > 0 && len(sets) == 0 {
				// Nothing to answer with and nothing safe to invent:
				// the prompt's rule, or a board that moved under it,
				// accepts no set this enumerator could find. Say so in
				// the default branch's words rather than hand the seat
				// an empty list in silence.
				slog.Warn("legal: no moves enumerated for a pending choice kind — the seat owing it has no legal move",
					"kind", c.Kind,
					"chooser", c.Chooser,
					"reason", reason,
				)
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

		case game.PendingChoiceMayCast:
			// #499. Cascade's "you may cast it without paying its mana
			// cost". ResolveMayCast refuses neither answer — a failing
			// branch is logged as an effect error, not returned — so
			// both are offered, and the decline is the always-legal
			// way out: it is the answer that commits the seat to
			// nothing.
			accept := true
			p := base()
			p.Apply = &accept
			e.addChoice(c, reason+": cast it", p)

			decline := false
			d := base()
			d.Apply = &decline
			e.addAlwaysLegalChoice(c, reason+": don't cast it", d)

		case game.PendingChoiceCreatureType:
			// #499. "As this enters, choose a creature type" — Cavern
			// of Souls, Door of Destinies, Adaptive Automaton. The
			// legal set is all ~345 CR 205.3m types, and offering
			// every one would bury a bot's list under names no deck
			// cares about.
			//
			// So the answers are the creature types of the creatures
			// this seat already controls, most common first, capped —
			// what a player naming a tribe for Door of Destinies does.
			// Read off the BATTLEFIELD only: it is public, so the
			// labels leak nothing the visibility guard would object
			// to, where the hand or library would.
			//
			// Every answer is always legal: ResolveCreatureTypeChoice
			// accepts any canonical type and treats a departed source
			// as "nowhere to land", not an error. A board with no
			// creatures still gets one answer, so the seat is never
			// handed an empty list — the wedge this case exists to end.
			for _, t := range e.creatureTypeAnswers() {
				p := base()
				p.CreatureType = t
				e.addAlwaysLegalChoice(c, reason+": "+t, p)
			}

		case game.PendingChoiceCardName:
			// #1210, CR 614.12. "As this enters, choose a card name"
			// — Pithing Needle, Phyrexian Revoker, Sorcerous
			// Spyglass. There is NO legal set (CR 201.2 admits any
			// card name at all), so this is not "offer the legal
			// answers, capped" the way the creature-type case is; it
			// is "offer the answers worth taking".
			//
			// Those are the names on the OPPONENTS' battlefield that
			// actually have an activated ability, most threatening
			// first, which is what a player points a Needle at. Read
			// off public zones only, so the labels leak nothing.
			//
			// Every answer is always legal: ResolveCardNameChoice
			// accepts any non-empty name and treats a departed source
			// as "nowhere to land", not an error. A board with
			// nothing worth naming still gets one answer, so the seat
			// is never handed an empty list — the wedge this case
			// exists to end.
			for _, n := range e.cardNameAnswers(c) {
				p := base()
				p.CardName = n
				e.addAlwaysLegalChoice(c, reason+": "+n, p)
			}

		case game.PendingChoiceLoopShortcut:
			// #804, CR 726. "<card> — <ability> has resolved N times
			// this turn. Resolve it K more times, then stop?" The
			// answer is a number, so the enumerator's job is to pick
			// the handful of numbers worth offering a policy — a human
			// types whatever they like into the client's field.
			//
			// The list is also where a bot-only table is made to
			// terminate, and it is here rather than in a policy on
			// purpose. A policy that ranks answers can rank "100 more"
			// top every time, and a random policy will eventually; the
			// loop then runs for as long as anything is willing to
			// keep saying yes. So the SECOND ask of a turn for the
			// same loop (LoopShortcutRepeat) offers one answer, stop,
			// and every policy at the table terminates because there
			// is nothing else to pick. A human is never in this list:
			// their client renders the number field from the prompt.
			//
			// First ask: 10, then 100, then stop. A bot takes the
			// first offer, which is why 10 leads. See docs/bot.md.
			//
			// #810 adds the second case that offers only stop: a loop
			// the chooser is DRIVING, by activating an ability of
			// their own permanent once per iteration. "Resolve it ten
			// more times" is an answer about a loop that runs itself;
			// for one the seat feeds by hand it buys nothing but ten
			// more turns of the crank, and the crank is the thing the
			// breaker is trying to stop. So a self-activated loop gets
			// the termination answer on the FIRST ask.
			offers := []int{
				game.DefaultLoopShortcutIterations,
				100,
				0,
			}
			if c.LoopShortcutRepeat || e.loopIsSelfActivated(c) {
				offers = []int{0}
			}
			for _, k := range offers {
				if k > game.MaxLoopShortcutIterations {
					continue
				}
				p := base()
				p.Iterations = k
				if k == 0 {
					// Stop is the always-legal way out: it commits the
					// seat to nothing and leaves the table exactly
					// where the breaker put it.
					e.addAlwaysLegalChoice(c, reason+": stop here", p)
					continue
				}
				e.addChoice(c, fmt.Sprintf("%s: resolve %d more times", reason, k), p)
			}

		default:
			// #499: a kind with no case above is enumerated NOTHING,
			// and `owed` is already true, so the seat gets an empty
			// move list — no answer and no pass either, because the
			// engine refuses pass_priority while a choice is open. A
			// bot runner reads that as "nothing to do" and sleeps
			// forever; a human client's legal_moves goes empty.
			//
			// The silence is the bug. A kind added to game/ ends up
			// here by default, and nothing anywhere says so until a
			// table wedges. Logging it does not un-wedge the seat, but
			// it turns "the game stopped" into a line naming the kind,
			// which is the difference between an hour and a minute.
			//
			// No live kind lands here today. The four #499 named all
			// have cases now; this branch is for the next one.
			slog.Warn("legal: no moves enumerated for a pending choice kind — the seat owing it has no legal move",
				"kind", c.Kind,
				"chooser", c.Chooser,
				"reason", reason,
			)
		}
	}
	return owed
}

// modeLabel renders a mode selection for the move list: the chosen
// bullets in order, so a bot's log says what it picked rather than
// which indexes it picked.
func modeLabel(c *game.PendingChoice, sel []int) string {
	if len(sel) == 0 {
		return ": none"
	}
	pos := make(map[int]int, len(c.ModeOptionIndex))
	for i, idx := range c.ModeOptionIndex {
		pos[idx] = i
	}
	out := ""
	for _, m := range sel {
		label := ""
		if i, ok := pos[m]; ok && i < len(c.ModeOptionLabel) {
			label = c.ModeOptionLabel[i]
		}
		if label == "" {
			label = "mode " + strconv.Itoa(m)
		}
		if out != "" {
			out += "; "
		}
		out += label
	}
	return ": " + out
}

// loopIsSelfActivated reports whether the loop a CR 726 shortcut
// prompt names is one its controller drives themselves, by activating
// an ability once per iteration, rather than a trigger loop that runs
// on its own once it starts (#810).
//
// Read off the ability the notice names: the prompt carries the
// source permanent and the stack label the repeating item announced
// with, and an ACTIVATED ability's stack label is its printed label
// (ActivateCatalogAbility), so a label that matches one of the
// source's activated abilities is an activation loop. A triggered
// ability's label never does — the catalog's convention for one is
// "<card> — <what happens>".
//
// No new state on the prompt for a fact the board already answers.
func (e *enumerator) loopIsSelfActivated(c *game.PendingChoice) bool {
	if c.Source == uuid.Nil || c.Reason == "" {
		return false
	}
	src := findBattlefield(e.g, c.Source)
	if src == nil {
		return false
	}
	for _, ab := range game.ActivatedAbilitiesForCard(*src) {
		if ab.Label == c.Reason {
			return true
		}
	}
	return false
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

// addAlwaysLegalChoice is addChoice for the one answer a choice kind
// can promise the engine will never refuse — see Move.AlwaysLegal. A
// kind that has such an answer should mark exactly one of them, and
// a kind that has none should mark nothing rather than guess.
func (e *enumerator) addAlwaysLegalChoice(c *game.PendingChoice, label string, p choiceParams) {
	e.add(Move{
		Type:        TypeResolveChoice,
		Player:      e.seat,
		Kind:        KindChoice,
		Label:       label,
		Source:      c.Source,
		Params:      mustJSON(p),
		AlwaysLegal: true,
	})
}

// filteredCombinations picks the answers a card-set prompt is offered
// — a search, or a choose-cards pick — out of `pool`: subsets of size
// lo..hi (lo raised to 1; the empty answer is the caller's to offer)
// that `allow` accepts, at most `limit` of them. It is combinations()
// with two differences that matter (#544).
//
// It FILTERS through `allow`, which is the engine's own acceptance
// check (SearchPickLegalLocked, ChooseCardsPickLegalLocked), so a pick
// the resolver would reject is never offered.
//
// And it SPREADS `limit` across pick sizes instead of spending it on
// the small ones first. That second half is not a refinement of the
// first; without it the filter is not enough. combinations() walks
// the smallest size to exhaustion before it reaches the next, so for a
// "search for up to two" prompt:
//
//   - with `limit` or more candidates the budget was gone before the
//     first pair, no pair was offered at all, and Myriad Landscape
//     quietly fetched one land off a card that prints two — for
//     every seat reading this list, bot or client;
//   - with fewer, exactly one or two pairs squeezed in, always the
//     first in pool order. Filtering an invalid one out of THAT list
//     leaves a legal list that still never offers a valid pair.
//
// A choose-cards prompt with a set-level rule is the same trap from
// the other side: "one creature card, or two cards" rejects every
// non-creature single. combinations() capped at `limit` and filtered
// afterwards hands out only singles once there are `limit` candidates,
// and filtering those leaves a hand with no creature in it an EMPTY
// list — no pair, no pass, the wedge itself.
//
// The budget has to reach the valid picks, not merely skip the
// invalid ones. Round-robin by size does that; ties in pool order
// are preserved within each size, so the first answer offered is
// still the first smallest subset in pool order.
func filteredCombinations(pool []uuid.UUID, lo, hi, limit int, allow func([]uuid.UUID) bool) [][]uuid.UUID {
	if hi > len(pool) {
		hi = len(pool)
	}
	if lo < 1 {
		lo = 1
	}
	if hi < lo || limit <= 0 {
		return nil
	}
	bySize := make([][][]uuid.UUID, hi+1)
	for k := lo; k <= hi; k++ {
		bySize[k] = allowedSubsets(pool, k, limit, allow)
	}
	out := make([][]uuid.UUID, 0, limit)
	for i := 0; len(out) < limit; i++ {
		before := len(out)
		for k := lo; k <= hi && len(out) < limit; k++ {
			if i < len(bySize[k]) {
				out = append(out, bySize[k][i])
			}
		}
		if len(out) == before {
			break
		}
	}
	return out
}

// subsetScanBudget bounds, per unit of `limit`, how many candidate
// subsets of one size allowedSubsets will test before giving up. A
// prompt whose rule rejects most sets (Myriad Landscape on a
// five-colour mana base) must not turn enumeration into a walk of
// every C(n,k).
const subsetScanBudget = 64

// allowedSubsets returns up to `limit` k-subsets of pool, in pool
// order, keeping only those `allow` accepts.
//
// The walk never extends a prefix that can no longer reach k cards:
// the loop stops while there are still enough cards left in the pool
// to finish the set. Without that, `scanned` (which counts only
// finished subsets) bounds nothing at the sizes where few subsets
// exist — k = n or n-1 — and the recursion visits every prefix, about
// 2^n nodes, all under the read lock. With it, every node leads to at
// least one finished subset, so the scan budget bounds the walk.
func allowedSubsets(pool []uuid.UUID, k, limit int, allow func([]uuid.UUID) bool) [][]uuid.UUID {
	if k > len(pool) || limit <= 0 {
		return nil
	}
	var out [][]uuid.UUID
	scanned, budget := 0, limit*subsetScanBudget
	var rec func(start int, cur []uuid.UUID)
	rec = func(start int, cur []uuid.UUID) {
		if len(out) >= limit || scanned >= budget {
			return
		}
		if len(cur) == k {
			scanned++
			if allow == nil || allow(cur) {
				out = append(out, append([]uuid.UUID(nil), cur...))
			}
			return
		}
		last := len(pool) - (k - len(cur)) // the last index that can still finish a k-set
		for i := start; i <= last && len(out) < limit && scanned < budget; i++ {
			rec(i+1, append(cur, pool[i]))
		}
	}
	rec(0, nil)
	return out
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

// colorAnswers orders a colour prompt's options by what the card said
// it will DO with the answer (#986). One ordering function, shared with
// the `choose_color` projection so a bot's move list and a human's
// buttons are in the same order; see color_order.go for the arms and
// for why this is a function rather than an Options hook.
func (e *enumerator) colorAnswers(c *game.PendingChoice) []string {
	return OrderColorOptionsLocked(e.g, e.seat, c.ColorOptions, c.ColorPurpose)
}

// creatureTypeAnswersCap bounds the creature types offered for one
// prompt. Enough to name every tribe a real Commander board runs, few
// enough that a random policy still lands on a sensible one.
const creatureTypeAnswersCap = 5

// creatureTypeFallback answers a creature-type prompt on a board with
// no creatures on it. Any canonical type is accepted; Human is the
// most-printed, so it is the likeliest to matter later.
const creatureTypeFallback = "Human"

// creatureTypeAnswers ranks the creature types among the creatures
// this seat controls on the battlefield, most common first and then
// alphabetically so the list is stable across calls.
//
// Changelings count for nothing here: every type is theirs, so they
// say nothing about which tribe the board is.
func (e *enumerator) creatureTypeAnswers() []string {
	counts := map[string]int{}
	for i := range e.g.Battlefield.Cards {
		c := &e.g.Battlefield.Cards[i]
		if c.Controller != e.seat || !c.IsCreature() || game.HasAllCreatureTypes(c) {
			continue
		}
		for _, t := range game.CreatureTypesOf(c) {
			if canon, ok := game.CanonicalCreatureType(t); ok {
				counts[canon]++
			}
		}
	}
	types := make([]string, 0, len(counts))
	for t := range counts {
		types = append(types, t)
	}
	sort.Slice(types, func(i, j int) bool {
		if counts[types[i]] != counts[types[j]] {
			return counts[types[i]] > counts[types[j]]
		}
		return types[i] < types[j]
	})
	if len(types) > creatureTypeAnswersCap {
		types = types[:creatureTypeAnswersCap]
	}
	if len(types) == 0 {
		return []string{creatureTypeFallback}
	}
	return types
}

// cardNameAnswersCap bounds the card names offered for one
// choose_card_name prompt. Small for the reason the creature-type cap
// is small, and smaller: the list is a judgement call ("what is worth
// naming") rather than a vocabulary, and a random policy taking the
// first offer should be taking a good one.
const cardNameAnswersCap = 5

// cardNameFallback answers a card-name prompt on a board with nothing
// worth naming. CR 201.2 accepts any card name, and a Needle that
// named the card it is stapled to restricts nothing — which is the
// honest answer to "there was nothing to point this at" and keeps the
// list non-empty, so the prompt can always be cleared.
const cardNameFallback = "Pithing Needle"

// cardNameAnswers ranks the card names worth naming for a
// choose_card_name prompt (#1210): the names on OPPONENTS'
// battlefields that actually have an activated ability, because that
// is the only thing a Pithing Needle or a Phyrexian Revoker does
// anything about.
//
// Ordered by how many such permanents carry the name (a table with
// two copies of one thing is saying something) and then
// alphabetically, so the list is stable across calls.
//
// PUBLIC zones only, and the battlefield is the only one that matters
// — a name in a hand is hidden, and a name in a graveyard has no
// ability anybody can activate.
func (e *enumerator) cardNameAnswers(c *game.PendingChoice) []string {
	chooser := e.seat
	if c != nil && c.Chooser != uuid.Nil {
		chooser = c.Chooser
	}
	counts := map[string]int{}
	for i := range e.g.Battlefield.Cards {
		card := &e.g.Battlefield.Cards[i]
		if card.Controller == chooser || card.FaceDown || card.Name == "" {
			continue
		}
		if len(game.ActivatedAbilitiesForCard(*card)) == 0 &&
			len(game.ManaAbilitiesForCard(*card)) == 0 {
			continue
		}
		counts[card.Name]++
	}
	names := make([]string, 0, len(counts))
	for n := range counts {
		names = append(names, n)
	}
	sort.Slice(names, func(i, j int) bool {
		if counts[names[i]] != counts[names[j]] {
			return counts[names[i]] > counts[names[j]]
		}
		return names[i] < names[j]
	})
	if len(names) > cardNameAnswersCap {
		names = names[:cardNameAnswersCap]
	}
	if len(names) == 0 {
		return []string{cardNameFallback}
	}
	return names
}

// choiceSeatName names a seat for a prompt label, falling back to
// "another player" for a seat the table no longer carries. Seat names
// are public (they are on every GameView), so a label may say one.
func choiceSeatName(g *game.Game, id uuid.UUID) string {
	if p := playerByID(g, id); p != nil && p.Name != "" {
		return p.Name
	}
	return "another player"
}

// cardSetPickPool is a card-set prompt's candidate list in the ORDER
// the enumerator should spend its expansion budget on — #1014's
// ordering hooks, reaching the three resolution-time picks (#1214).
//
// It matters because `allowedSubsets` walks the pool by index and
// stops at MaxExpansionPerSource: what the pool starts with is what
// survives the cap, and what survives the cap is the whole move list a
// policy gets to choose between. On a four-permanent board nothing is
// lost; on a real one, the set containing the card that matters simply
// is not offered, and no policy can pick a move it was never shown.
//
// WHICH HOOK, per kind — and the two hooks point OPPOSITE ways
// (legal.Options):
//
//   - their_permanents ranks somebody ELSE's board by what matters on
//     it, which is exactly Options.OrderTargets' question: "the part
//     of the board that matters" first, so the top of it reaches the
//     cap. reveal_pick is the same question about cards rather than
//     permanents — an opponent deciding which of your four tutored
//     cards you get is reading them the way they would read a board to
//     aim a removal spell.
//   - own_permanents ranks the seat's OWN permanents by what it would
//     miss least, which is Options.OrderCostFuel's question. Scapeshift
//     sacrificing any number of lands is a seat spending its own
//     board, and a policy handed OrderTargets there would keep its
//     worst land and sacrifice its best.
//
// choose_cards and untap_choice are deliberately left in engine order.
// Neither hook asks their question — one is usually a hand, and the
// other is CR 502.3's own determination, where the cap solver in the
// engine has already decided what is legal.
//
// A caller with neither hook set — every caller but a bot seat — gets
// the engine's own order back, byte for byte.
func (e *enumerator) cardSetPickPool(c *game.PendingChoice) []uuid.UUID {
	switch c.Kind {
	case game.PendingChoiceTheirPermanents, game.PendingChoiceRevealPick:
		return e.mostValuableFirst(c.ChooseCards)
	case game.PendingChoiceOwnPermanents:
		return e.cheapestFuelFirst(c.ChooseCards)
	}
	return c.ChooseCards
}

// mostValuableFirst is orderCandidates for a plain ID list: the same
// hook, the same stable descending sort, on a fresh slice because the
// pool comes from the engine and must not be reordered under it.
func (e *enumerator) mostValuableFirst(pool []uuid.UUID) []uuid.UUID {
	if e.opts.OrderTargets == nil || len(pool) < 2 {
		return pool
	}
	out := append([]uuid.UUID(nil), pool...)
	score := make(map[uuid.UUID]float64, len(out))
	for _, id := range out {
		score[id] = e.opts.OrderTargets(TargetCandidate{ID: id})
	}
	sort.SliceStable(out, func(i, j int) bool { return score[out[i]] > score[out[j]] })
	return out
}
