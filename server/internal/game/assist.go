package game

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

// assist.go — ADR 0150 §4 (#1857): assist, another player helping to
// pay a spell's generic mana.
//
// CR 702.132a: "Assist is a static ability that modifies the rules of
// paying for the spell with assist (see rules 601.2g-h). If the total
// cost to cast a spell with assist includes a generic mana component,
// before you activate mana abilities while casting it, you may choose
// another player. That player has a chance to activate mana abilities.
// Once that player chooses not to activate any more mana abilities, you
// have a chance to activate mana abilities. Before you begin to pay the
// total cost of the spell, the player you chose may pay for any amount
// of the generic mana in the spell's total cost."
//
// The cap is the generic part of the TOTAL cost (CR 601.2f: the mana or
// alternative cost, X, additional costs and increases, less reductions),
// so commander tax and a Thalia raise it and a reduction lowers it. The
// reminder text's "up to {3}" is the printed generic only.
//
// How the engine does it: the cast is PARKED before anything is paid,
// the shape #1397's commander-owner question built (cost_commander_
// choice.go). castSpellLocked validates the whole announcement, works
// out the total cost and its generic part G, and queues one pay_amount
// prompt to the helper (0..min(G, what they can make), resource mana,
// unit assist) carrying the announcement. Nothing has moved and nothing
// is paid. The helper's answer re-runs castSpellLocked with the amount
// attached, and the re-run re-validates everything, then plans BOTH
// payments before making either (CR 601.2h, "partial payments are not
// allowed"): the helper's n generic from their pool and auto-tapper,
// and the caster's cost less n. If either is short nothing is tapped or
// spent, the card stays where it was, and the caster gets a log line.
//
// The prompt's continuation rides the existing chooseValueResume frame
// rather than a frame of its own, so the closure ratchet
// (closure_fields.txt) gains no route; a table waiting on a helper is
// not a restore point, as no pay_amount table is.

// The refusals of a bad assist announcement. Each wraps
// ErrInvalidParam: the announcement is malformed, not unaffordable.
var (
	// ErrAssistNotOffered: assist_player on a spell without assist,
	// or on a forced cast, which pays no mana for anyone to help with.
	ErrAssistNotOffered = fmt.Errorf("%w: this spell has no assist", ErrInvalidParam)
	// ErrAssistPlayer: the caster, or a player who isn't in the game.
	ErrAssistPlayer = fmt.Errorf("%w: assist must name another player still in the game", ErrInvalidParam)
	// ErrAssistNoGeneric: the total cost has no generic part to help
	// with (CR 702.132a).
	ErrAssistNoGeneric = fmt.Errorf("%w: this spell's total cost has no generic mana to help with", ErrInvalidParam)
	// ErrAssistTooMuch: the helper's answer is more than the generic
	// the total cost has now.
	ErrAssistTooMuch = fmt.Errorf("%w: more help than the spell's generic mana", ErrInvalidParam)
)

// KeywordAssist is the canonical keyword token for assist (CR 702.132).
// A card has assist when it prints it: the importer stamps the token
// from Scryfall's keywords, and a catalog card declares it in
// PrintedKeywords.
const KeywordAssist = "assist"

// PayResourceMana is a pay_amount whose points are generic mana, paid
// from the chooser's pool and then their auto-tapper (ADR 0150 §4).
// Only assist asks one.
const PayResourceMana = "mana"

// PayAmountAssist is the unit of an assist prompt: each point is {1} of
// another player's spell.
const PayAmountAssist = "assist"

// AssistPrompt is the plain data an assist helper's pay_amount carries,
// for the view (the helper's header, the caster's waiting line) and the
// departure rule. Immutable once queued.
type AssistPrompt struct {
	// Caster is the player casting the spell; the prompt's Chooser is
	// the helper.
	Caster uuid.UUID
	// Spell is the card being cast, still where it was cast from.
	// SpellName and ManaCost are its name and its total cost as the
	// caster announced it ("{3}{B}"), carried because the card is in a
	// zone the helper may not see.
	Spell     uuid.UUID
	SpellName string
	ManaCost  string
	// Generic is the generic part of the spell's total cost, G: the
	// most anyone could help with. The prompt's Max is the smaller of G
	// and what the helper could make when asked.
	Generic int

	// spend is the cast's mana spend context, so an answer is checked
	// against what may pay for THIS spell ("spend only on creature
	// spells" mana may help cast Fan Favorite).
	spend ManaSpendContext
}

// assistAnswer is the helper's answer, attached to the parked
// announcement it re-runs. Unexported, and set only by the prompt's
// continuation, so no payload answers for another player.
type assistAnswer struct {
	helper uuid.UUID
	amount int
}

// lessFrom is `cost` with the helper's share of the generic taken off:
// the cost the caster still owes. X is settled first, so the share can
// come out of the mana announced for {X} too (it is generic, CR 107.3).
func (a *assistAnswer) lessFrom(cost ParsedCost, x int) ParsedCost {
	if a == nil || a.amount <= 0 {
		return cost
	}
	out := cost.SettleX(x)
	out.Generic -= a.amount
	if out.Generic < 0 {
		out.Generic = 0
	}
	return out
}

// HasAssist reports whether the card, as it would be cast, has assist.
func HasAssist(c *Card) bool { return HasKeyword(c, KeywordAssist) }

// validateAssistChoiceLocked is CR 702.132a's "choose another player",
// checked with the rest of CR 601.2b: only a spell with assist, only
// another player, and only one still in the game. Refused rather than
// ignored, the way a stray gift_opponent is: a helper sent for a spell
// that takes none is a client that thought it was asking for help.
//
// Caller must hold g.mu.
func (g *Game) validateAssistChoiceLocked(caster uuid.UUID, card *Card, params CastSpellParams) error {
	if params.AssistPlayer == uuid.Nil {
		return nil
	}
	if !HasAssist(card) || params.ForceCast {
		return ErrAssistNotOffered
	}
	if params.AssistPlayer == caster || !g.playerInGameLocked(params.AssistPlayer) {
		return ErrAssistPlayer
	}
	return nil
}

// AssistGenericLocked is the generic part of the total cost (CR 601.2f)
// `caster` would pay to cast `card` under `params`: what an assist helper
// may pay at most. False when the cost can't be read.
//
// Caller must hold g.mu.
func (g *Game) AssistGenericLocked(p *Player, card Card, params CastSpellParams) (int, bool) {
	params.assistPaid = nil
	cost, err := g.effectiveCostLocked(p, card, params)
	if err != nil {
		return 0, false
	}
	return cost.GenericWithX(params.XValue), true
}

// assistFundsLocked reports whether `helper` can pay `n` generic mana
// for a spell cast under `spend`, from their pool and then their
// auto-tapper, without tapping anything. It returns the plan that
// would do it.
//
// Caller must hold g.mu.
func (g *Game) assistFundsLocked(helper uuid.UUID, n int, spend ManaSpendContext) (tapPlan, ParsedCost, bool) {
	hp := g.playerByIDLocked(helper)
	if hp == nil || hp.Eliminated {
		return nil, ParsedCost{}, false
	}
	share := ParsedCost{Generic: n}
	if n <= 0 || hp.ManaPool.CanPayFor(share, 0, spend) {
		return nil, ParsedCost{}, n >= 0
	}
	plan, short, ok := g.autoTapTopUpLocked(helper, share, 0, spend, nil, 0)
	if !ok || !g.planFundsLocked(hp, plan, short, share, 0, spend) {
		return nil, ParsedCost{}, false
	}
	return plan, short, true
}

// assistCeilingLocked is the most generic mana `helper` could pay now,
// up to `generic`: the prompt's Max.
//
// Caller must hold g.mu.
func (g *Game) assistCeilingLocked(helper uuid.UUID, generic int, spend ManaSpendContext) int {
	for n := generic; n > 0; n-- {
		if _, _, ok := g.assistFundsLocked(helper, n, spend); ok {
			return n
		}
	}
	return 0
}

// parkForAssistLocked is the park: the announcement is valid, nothing
// has been paid, and the helper is asked how much of the generic they
// will pay. Reports parked=true when the prompt was queued; the caller
// returns without paying. A helper who could pay nothing is not asked,
// and the cast goes on unassisted.
//
// Caller must hold g.mu.
func (g *Game) parkForAssistLocked(p *Player, card Card, cardID uuid.UUID, params CastSpellParams) (bool, error) {
	generic, ok := g.AssistGenericLocked(p, card, params)
	if !ok {
		// The cost can't be read; the payment refuses the cast with the
		// parser's own error further down.
		return false, nil
	}
	if generic <= 0 {
		return false, ErrAssistNoGeneric
	}
	// One request per card: the same spell is already waiting on a
	// helper. (The table is not blocked for the caster's other casts,
	// as it isn't for #1397's park.)
	for _, c := range g.PendingChoices {
		if c != nil && c.Assist != nil && c.Assist.Spell == cardID {
			return false, ErrChoicePending
		}
	}
	spend := ManaSpendForCastParams(card, params)
	helper := params.AssistPlayer
	ceiling := g.assistCeilingLocked(helper, generic, spend)
	if ceiling == 0 {
		return false, nil
	}
	caster := p.ID
	spellName := card.Name
	announce := params
	id := g.QueueChoiceForEffect(PendingChoice{
		Kind:    PendingChoicePayAmount,
		Chooser: helper,
		Count:   1,
		Source:  cardID,
		Reason:  p.Name + " asks you to help pay for " + spellName,
		PayAmount: &PayAmountPrompt{
			Min: 0, Max: ceiling, Unit: PayAmountAssist, Resource: PayResourceMana,
		},
		Assist: &AssistPrompt{
			Caster:    caster,
			Spell:     cardID,
			SpellName: spellName,
			ManaCost:  g.assistCostTextLocked(p, card, params),
			Generic:   generic,
			spend:     spend,
		},
		chooseValueResume: &chooseValueFrame{then: func(g *Game, value string) error {
			// "" is the drop: the helper left (CR 800.4f), so they paid
			// nothing and the cast goes on unassisted.
			n, err := strconv.Atoi(value)
			if err != nil || n < 0 {
				n = 0
			}
			return g.resumeAssistedCastLocked(caster, cardID, spellName, announce, helper, n)
		}},
	})
	if id == uuid.Nil {
		return false, nil
	}
	return true, nil
}

// assistCostTextLocked is the spell's total cost as announced, for the
// helper's header.
func (g *Game) assistCostTextLocked(p *Player, card Card, params CastSpellParams) string {
	params.assistPaid = nil
	cost, err := g.effectiveCostLocked(p, card, params)
	if err != nil {
		return card.ManaCost
	}
	return cost.SettleX(params.XValue).String()
}

// resumeAssistedCastLocked makes the parked announcement again with the
// helper's answer attached. The re-run validates everything from
// scratch, and a refusal is the caster's log line rather than an error
// handed to the helper, whose answer was fine.
//
// Caller must hold g.mu.
func (g *Game) resumeAssistedCastLocked(caster, cardID uuid.UUID, spellName string, params CastSpellParams, helper uuid.UUID, n int) error {
	if !g.playerInGameLocked(caster) {
		return nil
	}
	params.assistPaid = &assistAnswer{helper: helper, amount: n}
	if n == 0 {
		// Nothing helped: an ordinary cast, which a departed helper
		// cannot refuse.
		params.AssistPlayer = uuid.Nil
	}
	if err := g.castSpellLocked(caster, cardID, params); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Actor:    caster,
			Source:   cardID,
			ErrorMsg: spellName + " couldn't be cast: " + err.Error(),
		})
		return nil
	}
	g.noteActionTakenLocked(caster)
	return nil
}

// assistCharge is a validated helper payment, ready to be made.
type assistCharge struct {
	helper *Player
	amount int
	plan   tapPlan
	short  ParsedCost
	spend  ManaSpendContext
}

// prepareAssistLocked is the re-run's first half of CR 601.2h: it plans
// the helper's share and checks the caster's remaining cost is payable,
// before anything is tapped or spent on either side. Nil, nil when no
// helper is paying.
//
// Caller must hold g.mu.
func (g *Game) prepareAssistLocked(p *Player, card Card, params CastSpellParams) (*assistCharge, error) {
	a := params.assistPaid
	if a == nil || a.amount <= 0 {
		return nil, nil
	}
	generic, ok := g.AssistGenericLocked(p, card, params)
	if !ok || a.amount > generic {
		return nil, ErrAssistTooMuch
	}
	hp := g.playerByIDLocked(a.helper)
	spend := ManaSpendForCastParams(card, params)
	plan, short, ok := g.assistFundsLocked(a.helper, a.amount, spend)
	if !ok || hp == nil {
		name := "the helper"
		if hp != nil {
			name = hp.Name
		}
		return nil, &AssistShortError{Helper: name, Amount: a.amount}
	}
	if err := g.castManaShortLocked(p, card, params); err != nil {
		return nil, err
	}
	return &assistCharge{helper: hp, amount: a.amount, plan: plan, short: short, spend: spend}, nil
}

// castManaShortLocked answers, without tapping or spending anything,
// whether the caster's mana payment would be refused: the same three
// outcomes applyAutoTapLocked and applyCastCostLocked reach, read
// rather than made.
//
// Caller must hold g.mu.
func (g *Game) castManaShortLocked(p *Player, card Card, params CastSpellParams) error {
	cost, err := g.effectiveCostLocked(p, card, params)
	if err != nil {
		return nil // refused by the payment itself, before anything is spent
	}
	spend := ManaSpendForCastParams(card, params)
	cost = g.costAsPaidByLocked(p.ID, spend, cost, params.XValue)
	cost, _, err = g.strikePhyrexianLifeLocked(p, card.Name, cost, spend, params.PhyrexianLife, nil)
	if err != nil {
		return nil
	}
	if p.ManaPool.CanPayFor(cost, params.XValue, spend) {
		return nil
	}
	if params.AutoTap {
		excluded := g.castAutoTapExclusionsLocked(p.ID, card, params)
		plan, short, ok := g.autoTapTopUpLocked(p.ID, cost, params.XValue, spend, excluded, WantedManaSourcesFor(card))
		if ok && g.planFundsLocked(p, plan, short, cost, params.XValue, spend) {
			return nil
		}
		return &InsufficientManaError{Missing: p.ManaPool.MissingFor(cost, params.XValue, spend)}
	}
	if !params.Strict {
		return nil // a permissive cast is paid on paper
	}
	return &InsufficientManaError{Missing: p.ManaPool.MissingFor(cost, params.XValue, spend)}
}

// payAssistLocked makes the helper's planned payment: their auto-tap
// plan, then the n generic from their pool, in the cast's spend
// context. Returns the mana spent, for the spell's payment record.
//
// Caller must hold g.mu.
func (g *Game) payAssistLocked(c *assistCharge, cardID uuid.UUID) ([]ManaToken, error) {
	if c == nil {
		return nil, nil
	}
	if len(c.plan) > 0 {
		g.materializePlanInLocked(c.helper, c.plan, c.short, c.spend)
	}
	share := ParsedCost{Generic: c.amount}
	spent, ok := c.helper.ManaPool.SpendManaForWith(share, 0, c.spend, SpendPreserveColors)
	if !ok {
		return nil, &AssistShortError{Helper: c.helper.Name, Amount: c.amount}
	}
	g.EmitEvent(manaSpentEvent(c.helper.ID, cardID, spent))
	return spent, nil
}

// AssistShortError is the re-run's refusal when the helper can no
// longer fund what they offered.
type AssistShortError struct {
	Helper string
	Amount int
}

func (e *AssistShortError) Error() string {
	return e.Helper + " can no longer pay {" + strconv.Itoa(e.Amount) + "}"
}

// Unwrap lets errors.Is(err, ErrInsufficientMana) match.
func (e *AssistShortError) Unwrap() error { return ErrInsufficientMana }

// AssistPlayersLocked is who `caster` may ask to help pay: every other
// player still in the game (CR 702.132a), in seat order.
//
// Caller must hold g.mu.
func (g *Game) AssistPlayersLocked(caster uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, s := range g.Seats {
		if s == nil || s.ID == caster || s.Eliminated {
			continue
		}
		out = append(out, s.ID)
	}
	return out
}

// AssistOfferLocked is what the cast surface offers `caster` for a
// spell with assist cast out of `zone` (ADR 0150 §5): the generic part
// of its total cost for the default announcement (no X, no optional
// cost), which the picker reads as "up to {N}", and who may be asked.
// False for a card without assist, or with no generic to help with.
//
// Caller must hold g.mu.
func (g *Game) AssistOfferLocked(caster uuid.UUID, card Card, zone ZoneKind) (int, []uuid.UUID, bool) {
	if !HasAssist(&card) {
		return 0, nil, false
	}
	p := g.playerByIDLocked(caster)
	if p == nil {
		return 0, nil, false
	}
	generic, ok := g.AssistGenericLocked(p, card, CastSpellParams{FromZone: string(zone)})
	if !ok || generic <= 0 {
		return 0, nil, false
	}
	return generic, g.AssistPlayersLocked(caster), true
}

// assistRequestForLocked is the open assist prompt `caster` is waiting
// on (the one with `choiceID`, when it is not uuid.Nil), or -1, nil.
// A caster who leaves the game has theirs withdrawn by the departure
// sweep (dropChoicesForPlayerLocked).
//
// Caller must hold g.mu.
func (g *Game) assistRequestForLocked(caster, choiceID uuid.UUID) (int, *PendingChoice) {
	for i, c := range g.PendingChoices {
		if c == nil || c.Assist == nil || c.Assist.Caster != caster {
			continue
		}
		if choiceID != uuid.Nil && c.ID != choiceID {
			continue
		}
		return i, c
	}
	return -1, nil
}

// CancelAssistRequest withdraws `caster`'s parked request for help
// (ADR 0150 owner decision 5): the helper's prompt goes, and since
// nothing was moved or paid the cast simply was not made. `choiceID`
// may be uuid.Nil for "my open request".
//
// Caller must NOT hold g.mu.
func (g *Game) CancelAssistRequest(caster, choiceID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx, c := g.assistRequestForLocked(caster, choiceID)
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	g.dequeueChoiceLocked(idx)
	g.EmitEvent(Event{
		Kind:   EventPendingChoiceDropped,
		Actor:  c.Chooser,
		Source: c.Source,
		Label:  string(c.Kind),
	})
	return nil
}

// cloneAssistPrompt copies an assist prompt for the undo clone.
func cloneAssistPrompt(a *AssistPrompt) *AssistPrompt {
	if a == nil {
		return nil
	}
	out := *a
	return &out
}
