package game

import (
	"slices"
	"strings"

	"github.com/google/uuid"
)

// combat_cost.go — #6190: a NON-MANA cost to attack or to block.
//
// CR 508.1h and CR 509.1d both say the same thing about what a cost to
// attack or block may be: "Costs may include paying mana, tapping
// permanents, sacrificing permanents, discarding cards, and so on."
// Until this file the engine's two taxes priced mana and life only
// (attack_tax.go, block_tax.go; #1063, #5375, #5869), and the attack
// tax's own comment named the gap: "A Sacrifice sibling would be one
// more field and one more branch in the payer."
//
// The cards:
//
//	Exalted Dragon      "can't attack unless you sacrifice a land"
//	Leviathan           "can't attack unless you sacrifice two Islands"
//	Flooded Woodlands   "Green creatures can't attack unless their
//	                     controller sacrifices a land of their choice for
//	                     each green creature they control that's attacking"
//	Reclamation         the same for black creatures
//	Hollow Warrior      "can't attack or block unless you tap an untapped
//	                     creature you control not declared as an attacking
//	                     or blocking creature this combat"
//	Floodtide Serpent   "can't attack unless you return an enchantment you
//	                     control to its owner's hand"
//
// # Nothing new is invented about paying
//
// The engine already models every component these clauses ask for, as
// parts of an ability's cost, each built on the same five pieces: the
// struct, one options walk, one payability predicate, one validator,
// one payer (ADR 0020 §2, #758, #1213). A cost to attack reuses all
// five:
//
//	sacrifice  AbilityCost.SacrificeOther's *TargetSpec, through
//	           validateSacrificeCostLocked and payCostSacrificesLocked
//	tap        TapOthersCost, through validateTapOthersCostLocked and
//	           payTapOthersCostLocked
//	return     ReturnToHandCost, through validateReturnToHandCostLocked
//	           and payReturnToHandCostLocked
//
// So a card file writes the clause in the vocabulary it already writes
// an activated ability's cost in, the client's pickers already know how
// to render the candidates, and a sacrifice paid to attack goes through
// exactly the battlefield exit a sacrifice paid to activate does — one
// EventSacrifice per permanent, dies-triggers, the simultaneous-exit
// batch.
//
// # Who chooses, and when
//
// The permanents are the PLAYER'S choice, so they ride the declaration
// the way an additional cost's sacrifice rides a cast: named on the
// action (CombatCostPayment, the sacrifice_ids / tap_ids / return_ids
// trio the wire already speaks for a cast and an activation), validated
// before anything is staged, and paid by the verb that declares. Never
// chosen by the engine: "sacrifice a land" is a decision no sandbox may
// make for a player.
//
// Which also settles the requirement rules for free. CR 508.1d and
// CR 509.1c: "If a creature can't attack/block unless a player pays a
// cost, that player is not required to pay that cost" — and a cost with
// a non-mana half is not free, so AttackTaxPrice.IsFree and
// BlockTaxPrice.IsFree answer false and the existing requirement
// searches already leave such a creature out.
//
// # What is deliberately not here
//
//   - A cost to attack paid by DISCARDING. CR 508.1h names it and no
//     card in the S82 corpus prints it; DiscardCost is the fourth
//     component and would be a fourth field, not a new shape.
//   - The CR 903.9 question for a commander sacrificed to pay. The
//     payment is one indivisible step (CR 601.2h, 602.2b) and a
//     declaration must not pause on a prompt, so the sacrifice is paid
//     "unasked", exactly as the auto-tapper's is (#1397,
//     payAbilityCostSacrificesLocked's `answers` nil). No card in the
//     family sacrifices anything a commander can be: the clauses name
//     lands, and a returned permanent never reaches CR 903.9.

// CombatCost is the non-mana half of a cost to attack or block, FOR ONE
// creature. Plain data built from the existing cost components, so a
// tax stays a struct a card file writes as a literal and nothing here
// is reachable from Game as a func field.
//
// The zero value demands nothing, so every pricer and payer can ask
// without a guard.
type CombatCost struct {
	// Sacrifice is "sacrifice a land" (Exalted Dragon), "sacrifice two
	// Islands" (Leviathan), "sacrifices a land of their choice for each
	// green creature they control that's attacking" (Flooded Woodlands,
	// one land per taxed attacker, because the price is summed per
	// attacking creature).
	//
	// The clause's Min == Max is how many, exactly as
	// AbilityCost.SacrificeOther's is (#747).
	Sacrifice *TargetSpec

	// Tap is "tap an untapped creature you control not declared as an
	// attacking or blocking creature this combat" (Hollow Warrior).
	// TapOthersCost, not the {T} symbol: summoning sickness does not
	// apply to a permanent tapped to pay a cost (CR 302.6, 602.5a).
	Tap *TapOthersCost

	// Return is "return an enchantment you control to its owner's hand"
	// (Floodtide Serpent). To its OWNER's hand (CR 109.5), which the
	// shared payer already gets right.
	Return *ReturnToHandCost

	// Label is the clause as printed, without the verb — "a land", "two
	// Islands", "an untapped creature you control not declared as an
	// attacking or blocking creature this combat". It is what the
	// player reads in the picker and in the refusal, so it is written
	// for a player and never in engine vocabulary.
	Label string
}

// Empty reports that this cost asks for nothing, which is every tax
// before #6190 and every tax that prices mana or life alone.
func (c CombatCost) Empty() bool {
	return c.Sacrifice == nil && c.Tap.Empty() && c.Return.Empty()
}

// CombatCostKind is which component one demand is, for the wire, the
// picker and the payment's bookkeeping. An on-disk identity only in the
// sense the view's strings are: never persisted in a snapshot.
type CombatCostKind string

const (
	// CombatCostSacrifice is "sacrifice <permanents>".
	CombatCostSacrifice CombatCostKind = "sacrifice"
	// CombatCostTap is "tap <permanents>".
	CombatCostTap CombatCostKind = "tap"
	// CombatCostReturn is "return <permanents> to its owner's hand".
	CombatCostReturn CombatCostKind = "return"
)

// CombatCostLine is one non-mana cost one permanent charges one
// attacking or blocking creature: the per-source breakdown the price
// carries, the refusal names and the view hands the client.
type CombatCostLine struct {
	// Source is the permanent with the clause, and Label that clause.
	Source uuid.UUID
	Label  string

	// Payer is the player who owes it: the attacking or blocking
	// creature's controller ("their controller"), never the source's.
	Payer uuid.UUID

	// Creature is the attacking or blocking creature the cost is for,
	// so a declaration of three creatures under Flooded Woodlands shows
	// three lines.
	Creature uuid.UUID

	// Cost is what must be paid for that one creature.
	Cost CombatCost
}

// combatCostSlot is ONE permanent a payment owes: a demand of one kind,
// with the component it was read from, flattened out of the lines so a
// payment of several lines is matched as one set.
//
// A slot rather than a count because the demands of one declaration can
// differ: Exalted Dragon's any land and Leviathan's two Islands in one
// declaration are three slots, and an Island pays any of them while a
// Forest pays only the first. Matching the chosen permanents to the
// slots is what gets that right (combatCostMatch).
type combatCostSlot struct {
	kind     CombatCostKind
	source   uuid.UUID
	creature uuid.UUID
	label    string
	// Exactly one of these is set, each a one-permanent copy of the
	// component the line declared, so the shared validators judge a
	// single pick.
	sacrifice *TargetSpec
	tap       *TapOthersCost
	ret       *ReturnToHandCost
}

// combatCostSlots flattens `lines` into the permanents they owe, in
// line order then component order (sacrifice, tap, return).
func combatCostSlots(lines []CombatCostLine) []combatCostSlot {
	var out []combatCostSlot
	for _, l := range lines {
		if spec := l.Cost.Sacrifice; spec != nil {
			one := *spec
			one.Min, one.Max = 1, 1
			for range max(specCount(spec), 1) {
				s := one
				out = append(out, combatCostSlot{kind: CombatCostSacrifice, source: l.Source, creature: l.Creature, label: l.Cost.Label, sacrifice: &s})
			}
		}
		if tc := l.Cost.Tap; !tc.Empty() {
			one := *tc
			one.Count = 1
			for range max(tc.Count, 1) {
				s := one
				out = append(out, combatCostSlot{kind: CombatCostTap, source: l.Source, creature: l.Creature, label: l.Cost.Label, tap: &s})
			}
		}
		if rc := l.Cost.Return; !rc.Empty() {
			one := *rc
			one.Count = 1
			for range max(rc.Count, 1) {
				s := one
				out = append(out, combatCostSlot{kind: CombatCostReturn, source: l.Source, creature: l.Creature, label: l.Cost.Label, ret: &s})
			}
		}
	}
	return out
}

// specCount is how many permanents a sacrifice clause asks for: its
// Min == Max count, as AbilityCost.SacrificeOther's is read.
func specCount(spec *TargetSpec) int {
	if spec == nil {
		return 0
	}
	return max(spec.Min, 1)
}

// admits reports whether `id` can pay this one slot, by the SAME
// validator the matching component's own cost uses — so a permanent the
// engine accepts here is one the engine's payer accepts, and the view's
// candidate list, the refusal and the payment cannot disagree (#544).
//
// Caller must hold g.mu with fresh layers. Reads only.
func (s combatCostSlot) admits(g *Game, payer, id uuid.UUID) bool {
	ids := []uuid.UUID{id}
	switch s.kind {
	case CombatCostSacrifice:
		_, err := g.validateSacrificeCostLocked(payer, s.source, AbilityCost{SacrificeOther: s.sacrifice}, ids, 0)
		return err == nil
	case CombatCostTap:
		return g.validateTapOthersCostLocked(payer, s.source, s.tap, ids) == nil
	case CombatCostReturn:
		return g.validateReturnToHandCostLocked(payer, s.source, s.ret, ids) == nil
	}
	return false
}

// candidatesLocked lists the permanents `payer` controls that could pay
// this slot, in battlefield order: the picker's options, through the
// same `admits` the payment is judged by.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (s combatCostSlot) candidatesLocked(g *Game, payer uuid.UUID) []uuid.UUID {
	if g.Battlefield == nil {
		return nil
	}
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != payer {
			continue
		}
		if s.admits(g, payer, c.InstanceID) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// CombatCostPayment is the permanents a player names to pay the
// non-mana cost of a declaration: the same three id lists a cast and an
// activation already send for the same three components.
//
// The ORDER does not matter, and neither does which list an id would
// also fit: the payment is matched to the demands (combatCostMatch), so
// a player who names an Island and a Forest for "sacrifice a land" plus
// "sacrifice two Islands" is refused rather than silently assigned a
// payment they did not make.
type CombatCostPayment struct {
	// SacrificeIDs are the permanents sacrificed, TapIDs those tapped
	// and ReturnIDs those returned to their owners' hands.
	SacrificeIDs []uuid.UUID
	TapIDs       []uuid.UUID
	ReturnIDs    []uuid.UUID
}

// Empty reports that nothing was named.
func (p CombatCostPayment) Empty() bool {
	return len(p.SacrificeIDs) == 0 && len(p.TapIDs) == 0 && len(p.ReturnIDs) == 0
}

// ids is every permanent named, in kind order.
func (p CombatCostPayment) ids() []uuid.UUID {
	out := make([]uuid.UUID, 0, len(p.SacrificeIDs)+len(p.TapIDs)+len(p.ReturnIDs))
	out = append(out, p.SacrificeIDs...)
	out = append(out, p.TapIDs...)
	out = append(out, p.ReturnIDs...)
	return out
}

// forKind is the ids named for one kind.
func (p CombatCostPayment) forKind(k CombatCostKind) []uuid.UUID {
	switch k {
	case CombatCostSacrifice:
		return p.SacrificeIDs
	case CombatCostTap:
		return p.TapIDs
	case CombatCostReturn:
		return p.ReturnIDs
	}
	return nil
}

// CombatCostUnpaidError is what a declaration refused for want of a
// non-mana cost returns: the clause that was not paid, in the words the
// card prints, for the client's toast.
//
// A struct rather than a sentinel for AttackTaxUnpaidError's reason: the
// player has to be told WHAT to sacrifice, tap or return, and a bare
// error would have meant a second call to find out.
type CombatCostUnpaidError struct {
	// Clause is the cost as printed ("a land", "two Islands").
	Clause string

	// Kind is which component it is, so the client opens the right
	// picker.
	Kind CombatCostKind

	// Creature is the attacking or blocking creature the cost is for,
	// and Source the permanent with the clause.
	Creature uuid.UUID
	Source   uuid.UUID

	// Err is the shared validator's own refusal (ErrInvalidParam for a
	// count, ErrIllegalTarget for a permanent the clause does not name),
	// or nil when nothing was named at all.
	Err error
}

func (e *CombatCostUnpaidError) Error() string { return "game: " + e.Sentence() }

// Sentence is the refusal in a player's words: "attacking or blocking
// here costs tapping an untapped creature you control, and it was not
// paid" — the verb from the kind, the clause as the card prints it.
//
// It names neither creature: the error carries both ids, and the client
// has the cards. Which half of combat it was is the action the player
// just sent, so the sentence says "attacking or blocking" rather than
// guessing.
func (e *CombatCostUnpaidError) Sentence() string {
	clause := e.Clause
	if clause == "" {
		clause = "a permanent you control"
	}
	what := "sacrificing " + clause
	switch e.Kind {
	case CombatCostTap:
		what = "tapping " + clause
	case CombatCostReturn:
		what = "returning " + clause + " to its owner's hand"
	}
	return "attacking or blocking here costs " + what + ", and it was not paid"
}

// Unwrap lets errors.Is(err, ErrCombatCostUnpaid) match and errors.As
// reach the component validator's own error.
func (e *CombatCostUnpaidError) Unwrap() []error {
	if e.Err == nil {
		return []error{ErrCombatCostUnpaid}
	}
	return []error{ErrCombatCostUnpaid, e.Err}
}

// combatCostRefusalLocked is the one validator: whether `pay` pays
// `lines` exactly, as the payment that is about to happen. nil means it
// does.
//
// What it enforces, in the order the errors matter:
//
//   - ids sent for a declaration that owes nothing are REFUSED rather
//     than ignored, as an activation's sacrifice_ids are: a client that
//     sends them is confused about what it is declaring.
//   - exactly as many permanents per kind as the demands ask for. A cost
//     is not partially payable (CR 118.3).
//   - no permanent named twice, in one list or across them: one
//     permanent pays one slot, and a permanent cannot be both
//     sacrificed and tapped.
//   - a MATCHING of the named permanents onto the slots. Each slot is
//     judged by its component's own validator, so every rule those
//     already enforce — on the battlefield, controlled by the payer,
//     untapped for a tap, not "can't be sacrificed", matched by the
//     clause without the targeting gate — is enforced once, here.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) combatCostRefusalLocked(payer uuid.UUID, lines []CombatCostLine, pay CombatCostPayment) *CombatCostUnpaidError {
	slots := combatCostSlots(lines)
	if len(slots) == 0 {
		if !pay.Empty() {
			return &CombatCostUnpaidError{Err: ErrInvalidParam}
		}
		return nil
	}
	named := pay.ids()
	seen := make(map[uuid.UUID]bool, len(named))
	for _, id := range named {
		if seen[id] {
			return combatCostError(slots[0], ErrInvalidParam)
		}
		seen[id] = true
	}
	for _, k := range []CombatCostKind{CombatCostSacrifice, CombatCostTap, CombatCostReturn} {
		want := 0
		var first combatCostSlot
		for _, s := range slots {
			if s.kind == k {
				if want == 0 {
					first = s
				}
				want++
			}
		}
		got := len(pay.forKind(k))
		if want != got {
			if want == 0 {
				return &CombatCostUnpaidError{Kind: k, Err: ErrInvalidParam}
			}
			return combatCostError(first, ErrInvalidParam)
		}
	}
	for _, k := range []CombatCostKind{CombatCostSacrifice, CombatCostTap, CombatCostReturn} {
		kindSlots := make([]combatCostSlot, 0, len(slots))
		for _, s := range slots {
			if s.kind == k {
				kindSlots = append(kindSlots, s)
			}
		}
		if len(kindSlots) == 0 {
			continue
		}
		ids := pay.forKind(k)
		if bad, ok := combatCostMatch(g, payer, kindSlots, ids); !ok {
			return combatCostError(kindSlots[0], badCombatCostErr(bad))
		}
	}
	return nil
}

// badCombatCostErr is the refusal a failed matching reports: a named
// permanent no slot accepts is an illegal choice, and a set that is the
// right size with every permanent legal somewhere but no assignment is
// an invalid payment.
func badCombatCostErr(bad bool) error {
	if bad {
		return ErrIllegalTarget
	}
	return ErrInvalidParam
}

// combatCostError wraps a slot's refusal as the player's sentence.
func combatCostError(s combatCostSlot, err error) *CombatCostUnpaidError {
	return &CombatCostUnpaidError{
		Clause: s.label, Kind: s.kind, Source: s.source, Creature: s.creature, Err: err,
	}
}

// combatCostMatch reports whether every id can be assigned to a slot of
// its kind, one each: Kuhn's augmenting-path matching over a handful of
// slots, which is exact where a greedy pass is not (an Island named for
// "sacrifice a land" while two Islands are owed).
//
// `bad` is set when some id fits no slot at all, which is the refusal a
// player can act on ("that permanent doesn't pay this").
//
// Caller must hold g.mu with fresh layers. Reads only.
func combatCostMatch(g *Game, payer uuid.UUID, slots []combatCostSlot, ids []uuid.UUID) (bad, ok bool) {
	if len(ids) != len(slots) {
		return false, false
	}
	admits := make([][]bool, len(ids))
	for i, id := range ids {
		admits[i] = make([]bool, len(slots))
		any := false
		for j, s := range slots {
			if s.admits(g, payer, id) {
				admits[i][j], any = true, true
			}
		}
		if !any {
			return true, false
		}
	}
	slotFor := make([]int, len(slots))
	for i := range slotFor {
		slotFor[i] = -1
	}
	var augment func(i int, seen []bool) bool
	augment = func(i int, seen []bool) bool {
		for j := range slots {
			if !admits[i][j] || seen[j] {
				continue
			}
			seen[j] = true
			if slotFor[j] == -1 || augment(slotFor[j], seen) {
				slotFor[j] = i
				return true
			}
		}
		return false
	}
	for i := range ids {
		if !augment(i, make([]bool, len(slots))) {
			return false, false
		}
	}
	return false, true
}

// combatCostPayableLocked reports whether `payer` could pay `lines` at
// all right now: a matching exists from the permanents they control
// onto the demands. It does not pay anything.
//
// The question the enumerator and the pricer's "could this declaration
// ever be legal" callers ask, and the one that keeps a declaration that
// can never be paid from being offered (#544).
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) combatCostPayableLocked(payer uuid.UUID, lines []CombatCostLine) bool {
	slots := combatCostSlots(lines)
	if len(slots) == 0 {
		return true
	}
	// Candidates per slot, then the same matching the payment is judged
	// by — so "two Islands and a land" is payable with two Islands and a
	// Forest and not with one Island and two Forests.
	cands := make([][]uuid.UUID, len(slots))
	pool := make([]uuid.UUID, 0, 8)
	for i, s := range slots {
		cands[i] = s.candidatesLocked(g, payer)
		if len(cands[i]) == 0 {
			return false
		}
		for _, id := range cands[i] {
			if !slices.Contains(pool, id) {
				pool = append(pool, id)
			}
		}
	}
	idFor := make([]int, len(pool))
	for i := range idFor {
		idFor[i] = -1
	}
	var augment func(j int, seen []bool) bool
	augment = func(j int, seen []bool) bool {
		for i, id := range pool {
			if seen[i] || !slices.Contains(cands[j], id) {
				continue
			}
			seen[i] = true
			if idFor[i] == -1 || augment(idFor[i], seen) {
				idFor[i] = j
				return true
			}
		}
		return false
	}
	for j := range slots {
		if !augment(j, make([]bool, len(pool))) {
			return false
		}
	}
	return true
}

// payCombatCostsLocked pays `pay` for `lines`, after
// combatCostRefusalLocked has accepted it: the taps first (nothing
// leaves the battlefield), then the returns, then the sacrifices, so a
// permanent that pays one component is still there while another's is
// read.
//
// Each component goes through its own shared payer, so a sacrifice is a
// sacrifice (EventSacrifice, dies-triggers, one simultaneous exit), a
// tap emits EventTapCard and a return takes the one battlefield exit to
// its OWNER's hand.
//
// Caller must hold g.mu in write mode, with the payment validated.
func (g *Game) payCombatCostsLocked(payer uuid.UUID, lines []CombatCostLine, pay CombatCostPayment) error {
	if len(lines) == 0 || pay.Empty() {
		return nil
	}
	g.payTapOthersCostLocked(payer, pay.TapIDs)
	for _, id := range pay.ReturnIDs {
		// Source uuid.UUID: the clause is the taxing permanent's, and
		// the return names no ability of its own.
		if _, err := g.payReturnToHandCostLocked(payer, uuid.Nil, []uuid.UUID{id}, nil); err != nil {
			return err
		}
	}
	// `answers` nil: the CR 903.9 question is not asked, the way the
	// auto-tapper's payment does not ask it (see the file comment).
	if err := g.payCostSacrificesLocked(pay.SacrificeIDs, nil); err != nil {
		return err
	}
	return nil
}

// combatCostDescribe is the non-mana half of a price as a player reads
// it: "sacrificing a land", "tapping an untapped creature you control",
// joined by " and " when a declaration owes several. "" for none.
func combatCostDescribe(lines []CombatCostLine) string {
	var parts []string
	for _, l := range lines {
		for _, p := range []struct {
			kind CombatCostKind
			on   bool
		}{
			{CombatCostSacrifice, l.Cost.Sacrifice != nil},
			{CombatCostTap, !l.Cost.Tap.Empty()},
			{CombatCostReturn, !l.Cost.Return.Empty()},
		} {
			if !p.on {
				continue
			}
			clause := l.Cost.Label
			if clause == "" {
				clause = "a permanent you control"
			}
			var s string
			switch p.kind {
			case CombatCostTap:
				s = "tapping " + clause
			case CombatCostReturn:
				s = "returning " + clause + " to its owner's hand"
			default:
				s = "sacrificing " + clause
			}
			if !slices.Contains(parts, s) {
				parts = append(parts, s)
			}
		}
	}
	return strings.Join(parts, " and ")
}

// CombatCostDemandView is one non-mana demand as the client needs it to
// open a picker: the clause, the kind, how many permanents, and the
// candidates the payer could name. Built by the view (ADR 0045 §6's
// rule that the client never re-derives a rule).
type CombatCostDemandView struct {
	// Kind is which picker to open, Clause the cost as printed, and
	// SourceName the permanent that charges it.
	Kind       CombatCostKind
	Clause     string
	SourceName string

	// Count is how many permanents this demand asks for, and Candidates
	// the ones the payer controls that could pay it, in battlefield
	// order.
	Count      int
	Candidates []uuid.UUID
}

// combatCostDemandsLocked collapses `lines` into the demands the client
// prompts for: one per (kind, clause, source), with the count summed and
// the candidates of the first slot of that group.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) combatCostDemandsLocked(payer uuid.UUID, lines []CombatCostLine) []CombatCostDemandView {
	slots := combatCostSlots(lines)
	if len(slots) == 0 {
		return nil
	}
	type key struct {
		kind   CombatCostKind
		clause string
		source uuid.UUID
	}
	var order []key
	byKey := map[key]*CombatCostDemandView{}
	for _, s := range slots {
		k := key{s.kind, s.label, s.source}
		d, ok := byKey[k]
		if !ok {
			name := ""
			if c := findBattlefieldCard(g, s.source); c != nil {
				name = c.Effective().Name
			}
			d = &CombatCostDemandView{
				Kind:       s.kind,
				Clause:     s.label,
				SourceName: name,
				Candidates: s.candidatesLocked(g, payer),
			}
			byKey[k] = d
			order = append(order, k)
		}
		d.Count++
	}
	out := make([]CombatCostDemandView, 0, len(order))
	for _, k := range order {
		out = append(out, *byKey[k])
	}
	return out
}

// AttackCostDemandsForEffect is the non-mana cost `attacker`'s
// controller must pay to declare it attacking `target` (#6190), as the
// view hands it to the client: the clause, the kind, how many
// permanents and which ones could pay. Nil when attacking costs no
// permanents, which is every creature at nearly every table.
//
// Priced through the declaration pricer, so what the client prompts for
// is exactly what the verb will demand (#544).
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) AttackCostDemandsForEffect(attacker *Card, target uuid.UUID) []CombatCostDemandView {
	if attacker == nil || target == uuid.Nil {
		return nil
	}
	price := g.priceAttackDeclarationLocked([]AttackDeclaration{{Attacker: attacker.InstanceID, Target: target}})
	return g.combatCostDemandsLocked(attacker.Controller, price.NonMana)
}

// BlockCostDemandsForEffect is the non-mana cost `blocker`'s controller
// must pay to keep it blocking when their declaration completes (#6190,
// CR 509.1d). Asked of one creature, so the dock can show the cost
// before the block is staged.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) BlockCostDemandsForEffect(blocker *Card) []CombatCostDemandView {
	if blocker == nil {
		return nil
	}
	price := g.priceBlockPlansLocked([]blockPlan{g.stagedBlockPlanLocked(blocker)})
	return g.combatCostDemandsLocked(blocker.Controller, price.NonMana)
}

// StagedBlockCostDemandsForEffect is the non-mana cost `seat` owes for
// the blocks they have staged (#6190): what finishing the declaration
// will demand, which is what the "Done blocking" button must pay.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) StagedBlockCostDemandsForEffect(seat uuid.UUID) []CombatCostDemandView {
	return g.combatCostDemandsLocked(seat, g.priceStagedBlocksLocked(seat).NonMana)
}

// CombatCostPayableForEffect reports whether `payer` could pay the
// non-mana demands of `lines` from the board they control: the
// enumerator's and the dock's "is this declaration possible at all".
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) CombatCostPayableForEffect(payer uuid.UUID, lines []CombatCostLine) bool {
	return g.combatCostPayableLocked(payer, lines)
}
