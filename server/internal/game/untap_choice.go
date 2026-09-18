package game

import "github.com/google/uuid"

// untap_choice.go — CR 502.3's FIRST sentence: "the active player
// determines which permanents they control will untap."
//
// untap.go answers the second one. The set of permanents that untap is
// computed there from the board: the active player's tapped permanents,
// less whatever an UntapStepRestriction or a next-untap marker holds
// back (ADR 0058), plus whatever an UntapStepPermission adds (#74).
// There is no player in that sentence, and for every card in the
// catalog before this file that was right — every effect in it is a
// flat "doesn't untap".
//
// Two printed families make the determination a real decision, and ADR
// 0070 is where they are decided:
//
//	"Players can't untap more than one land during their untap
//	 steps."                                        (Winter Orb, 9 cards)
//
// is an UntapCap: a ceiling of N over a matching set. Several compose,
// and a chosen set is legal iff EVERY cap is satisfied.
//
//	"You may choose not to untap this artifact during your untap
//	 step."                                 (Amber Prison, 45 cards)
//
// is an UntapOptOut: a per-object exemption from the fact that
// untapping is otherwise mandatory. The existence of this printed text
// is the proof that it is mandatory — CR 502.3's "normally, all of a
// player's permanents untap" is the default, and a cap only removes the
// excess.
//
// # The pause
//
// The untap step grants nobody priority (CR 502.4) and every step-entry
// path recurses straight through it inside one write lock. So a prompt
// here is a turn-based action stopping halfway, which the engine has
// done twice before and named once: #710 split finishStepEntryLocked
// out so a CR 616 ordering prompt could pause a step ENTRY, and #661
// gave the cleanup step exitCleanupStepLocked, ONE function that is
// that step's exit, called from the step-entry hook and from the
// discard resume. exitUntapStepLocked below is the third of that shape
// and the same rule applies: the two ways this step's turn-based
// actions can finish must not decide separately.
//
// Nothing untaps until the determination is complete. CR 502.3 is
// "determine, THEN untap them all simultaneously", so a paused step has
// untapped nothing and the whole set untaps in one loop from the
// continuation, through the same untapPermanentLocked primitive the
// unpaused path uses.

// PendingChoiceUntapChoice is CR 502.3's determination as a prompt:
// "choose which of these untap", addressed to the active player, over
// the permanents that are actually in question — the members of a cap
// that binds, and the permanents their controller may choose not to
// untap.
//
// It carries a choose-cards prompt's payload ({card_ids}), bounds
// (ChooseMin / ChooseMax) and continuation (chooseCardsResume), and is
// validated by the same checkChooseCardsPicksLocked. It is a separate
// KIND from PendingChoiceChooseCards for three reasons, of which the
// third decides it:
//
//   - The candidates are public. filterPendingChoices strips a
//     choose_cards prompt's options AND its bounds from every seat but
//     the chooser, because those candidates are usually a hand. These
//     are tapped permanents on the battlefield.
//   - The sentence is different. "Choose cards" is not "choose which of
//     these untap", and the kind is what the client renders from.
//   - The SIGN is inverted for a bot. The heuristic's choose_cards
//     branch scores an answer by what it does NOT name (#798) — every
//     choose_cards prompt over a bot's own cards is a discard or a
//     put-back. Naming a permanent here means UNTAPPING it, and scored
//     through that branch a Winter Orb answer falls to the flat
//     "no opinion" case: the bot takes the first set the enumerator
//     offers, which is battlefield order. That is precisely the
//     automatic pick ADR 0058's question 3 rejected for humans.
const PendingChoiceUntapChoice PendingChoiceKind = "untap_choice"

// UntapCap is one "players can't untap more than N <kind> during their
// untap steps" clause (CR 502.3) — Winter Orb, Static Orb, Winter Moon,
// Damping Field.
//
// Declared on effects.Spec.UntapCaps and bridged into this package by
// the CatalogUntapCaps hook, the way UntapStepPermission and
// UntapStepRestriction are. Both predicates run under g.mu held in
// write mode and MUST NOT call public locking mutators.
type UntapCap struct {
	// Applies reports whether this cap is live for `activePlayer`'s
	// untap step. Winter Orb and Static Orb print "as long as this
	// artifact is untapped"; Winter Moon prints no condition and is
	// always live. Same signature and argument order as
	// UntapStepPermission.AppliesTo.
	//
	// It is NOT where "whose untap step is this" is decided — see
	// untapChoicePlanLocked, which only ever consults a cap for
	// permanents the active player controls, because every printed
	// card in the family says "during THEIR untap steps" (ADR 0070
	// Decision 4). Nil means "always live".
	Applies func(g *Game, source *Card, activePlayer uuid.UUID) bool

	// Counts reports whether `target` is one of the permanents this
	// cap counts. Winter Orb: a land. Winter Moon: a nonbasic land.
	// Static Orb: every permanent. Type tests read the EFFECTIVE
	// characteristics, so the cap is read off fresh layers.
	//
	// Nil means "counts everything", which is Static Orb.
	Counts func(g *Game, source *Card, target *Card) bool

	// Max is the N in "no more than N". A cap of zero or less is not
	// a cap at all, it is a flat restriction, and belongs on
	// Spec.UntapStepRestrictions; such an entry is ignored here
	// rather than silently turned into one.
	Max int

	// Label names the clause for logs and debugging.
	Label string
}

// UntapOptOut declares "you may choose not to untap this during your
// untap step" (CR 502.3) — Rust Tick, Amber Prison, and 43 more.
//
// Same signature and argument order as UntapStepRestriction, so the
// same self / attached / filtered predicates plug in unchanged. Every
// printed card in the family is self-only today; the shape is the
// general one anyway, because the restriction it mirrors needed all
// three.
type UntapOptOut struct {
	// Optional reports whether `target` is a permanent its controller
	// may choose not to untap, given `source` declared the clause.
	Optional func(target *Card, g *Game, source *Card) bool
	// Label names the clause for logs and debugging.
	Label string
}

// CatalogUntapCaps returns the untap caps declared by the given oracle
// ID, or nil — which is every card but a handful. Nil hook means no
// catalog is wired and the untap step caps nothing, which is the
// pre-ADR-0070 behaviour and what the game package's own tests see
// unless they stub it.
var CatalogUntapCaps func(oracleID string) []UntapCap

// CatalogUntapOptOuts is the same for "you may choose not to untap".
var CatalogUntapOptOuts func(oracleID string) []UntapOptOut

// untapChoicePlan is one untap step's determination, frozen as DATA at
// the moment the prompt is queued.
//
// Data, not closures, for the reason ChooseCardsPrompt.Validate gives:
// the legality check runs on the submit path under the write lock AND
// inside legal.EnumerateFor under the READ lock, against whichever
// clone the enumerator was handed. A predicate that could reach the
// game would put the whole *ForEffect mutation surface one call away
// from a read-locked caller; a function of the picks cannot mutate
// anything at all.
type untapChoicePlan struct {
	// settled are the permanents that untap with no question asked —
	// everything in the CR 502.3 set that no cap counts and no clause
	// lets its controller hold back, including permanents another
	// player untaps through an UntapStepPermission.
	settled []uuid.UUID

	// candidates are the permanents the prompt asks about, in
	// battlefield order.
	candidates []uuid.UUID

	// optional[i] is true when candidates[i] is one its controller may
	// choose not to untap. Such a candidate is exempt from the
	// maximality clause in legal().
	optional []bool

	// caps are the live ceilings. members[c][i] is true when
	// candidates[i] is counted by caps[c].
	capMax  []int
	members [][]bool
}

// legal is THE answer to "would this set be accepted", and the only
// copy of it. ADR 0070 Decision 3, both clauses:
//
//	for every cap i:  |picks ∩ counted(i)| <= Max(i)
//	for every candidate c not picked and not opt-out-able:
//	                  picks ∪ {c} violates some cap
//
// The second clause is what stops a player declining an untap they
// have no printed permission to decline. Untapping is mandatory; a cap
// removes the excess and nothing else.
//
// Picks are assumed to be distinct members of `candidates` — the
// caller (checkChooseCardsPicksLocked) has already checked membership,
// duplicates and the bounds.
func (p *untapChoicePlan) legal(picks map[uuid.UUID]bool) bool {
	counts := make([]int, len(p.capMax))
	for i, id := range p.candidates {
		if !picks[id] {
			continue
		}
		for c := range p.capMax {
			if p.members[c][i] {
				counts[c]++
			}
		}
	}
	for c, max := range p.capMax {
		if counts[c] > max {
			return false
		}
	}
	// Maximality: every mandatory untap that COULD still happen must.
	for i, id := range p.candidates {
		if picks[id] || p.optional[i] {
			continue
		}
		blocked := false
		for c, max := range p.capMax {
			if p.members[c][i] && counts[c] >= max {
				blocked = true
				break
			}
		}
		if !blocked {
			return false
		}
	}
	return true
}

// bounds are the count guards the prompt carries.
//
// max is a genuine UPPER bound, so the count guard can never refuse a
// set legal() would accept: a legal set has at most Max(c) members
// inside cap c and at most everything outside it. It is exact for the
// nested families real cards print (nonbasic lands ⊂ lands ⊂
// permanents) and conservative otherwise.
//
// min is 1 whenever any candidate is NOT opt-out-able, and 0
// otherwise, and it is doing one specific job: ChooseCardsPrompt
// deliberately does not run Validate on an empty pick, because "choose
// nothing" is the answer the engine can never refuse and the one the
// enumerator marks AlwaysLegal. So the empty set has to be excluded by
// the floor or not at all. Between the floor and the ceiling, legal()
// decides — a floor equal to the true minimum legal size is the same
// set-packing problem as the true maximum, and buys only a more
// accurate submit button (ADR 0070, alternatives).
func (p *untapChoicePlan) bounds() (min, max int) {
	max = len(p.candidates)
	for c, capMax := range p.capMax {
		outside := 0
		for i := range p.candidates {
			if !p.members[c][i] {
				outside++
			}
		}
		if n := capMax + outside; n < max {
			max = n
		}
	}
	if max < 0 {
		max = 0
	}
	for _, opt := range p.optional {
		if !opt {
			min = 1
			break
		}
	}
	if min > max {
		min = max
	}
	return min, max
}

// boundUntapCap / boundUntapOptOut pair a declared clause with the
// battlefield card that declared it, as boundUntapPermission does.
// Pointers into g.Battlefield.Cards, valid only for the duration of
// the write-locked step that built them.
type boundUntapCap struct {
	cap    UntapCap
	source *Card
}

type boundUntapOptOut struct {
	optOut UntapOptOut
	source *Card
}

// activeUntapCapsLocked gathers the caps that are live for
// `activePlayer`'s untap step. One walk of the battlefield, through
// CatalogAbilityKey so a Winter Orb that has lost its abilities caps
// nothing. Caller must hold g.mu in write mode.
func (g *Game) activeUntapCapsLocked(activePlayer uuid.UUID) []boundUntapCap {
	if g.Battlefield == nil || CatalogUntapCaps == nil {
		return nil
	}
	var out []boundUntapCap
	for i := range g.Battlefield.Cards {
		source := &g.Battlefield.Cards[i]
		key := CatalogAbilityKey(*source)
		if key == "" {
			continue
		}
		for _, c := range CatalogUntapCaps(key) {
			if c.Max <= 0 {
				continue
			}
			if c.Applies != nil && !c.Applies(g, source, activePlayer) {
				continue
			}
			out = append(out, boundUntapCap{cap: c, source: source})
		}
	}
	return out
}

// activeUntapOptOutsLocked is the same for the "you may choose not to
// untap" clauses. Caller must hold g.mu in write mode.
func (g *Game) activeUntapOptOutsLocked() []boundUntapOptOut {
	if g.Battlefield == nil || CatalogUntapOptOuts == nil {
		return nil
	}
	var out []boundUntapOptOut
	for i := range g.Battlefield.Cards {
		source := &g.Battlefield.Cards[i]
		key := CatalogAbilityKey(*source)
		if key == "" {
			continue
		}
		for _, o := range CatalogUntapOptOuts(key) {
			if o.Optional != nil {
				out = append(out, boundUntapOptOut{optOut: o, source: source})
			}
		}
	}
	return out
}

// untapChoicePlanLocked splits the CR 502.3 set into the permanents
// that untap without a question and the permanents the active player
// is asked about, and freezes the rule the answer has to satisfy.
//
// `set` is untapStepSetLocked's answer, so anything an
// UntapStepRestriction or a next-untap marker holds back is already
// gone and can never be offered as one of the N choices — the sentence
// ADR 0058 Decision 8 wrote for this ADR.
//
// Returns nil when there is nothing to decide, which is almost every
// untap step of almost every game. A cap that does not BIND (Winter
// Orb with one tapped land) puts nothing in the prompt: the ceiling is
// only a decision when there are more eligible permanents than it
// allows.
//
// Caller must hold g.mu in write mode.
func (g *Game) untapChoicePlanLocked(activePlayer uuid.UUID, set []uuid.UUID) *untapChoicePlan {
	if len(set) == 0 || g.Battlefield == nil {
		return nil
	}
	caps := g.activeUntapCapsLocked(activePlayer)
	optOuts := g.activeUntapOptOutsLocked()
	if len(caps) == 0 && len(optOuts) == 0 {
		return nil
	}
	// Only the active player's OWN permanents are in question. Every
	// printed card in both families is scoped to a player's own untap
	// step ("during their untap steps", "during your untap step"), so
	// a Seedborn Muse untapping on somebody else's turn is not in
	// one — that step is not the Muse controller's. Same argument ADR
	// 0058 Decision 1 makes for restrictions.
	type entry struct {
		id       uuid.UUID
		card     *Card
		optional bool
		inCap    []bool
	}
	var own []entry
	var settled []uuid.UUID
	for _, id := range set {
		c := g.findBattlefieldCardLocked(id)
		if c == nil || c.Controller != activePlayer {
			settled = append(settled, id)
			continue
		}
		e := entry{id: id, card: c, inCap: make([]bool, len(caps))}
		for i, bc := range caps {
			e.inCap[i] = bc.cap.Counts == nil || bc.cap.Counts(g, bc.source, c)
		}
		for _, bo := range optOuts {
			if bo.optOut.Optional(c, g, bo.source) {
				e.optional = true
				break
			}
		}
		own = append(own, e)
	}
	// A cap binds only when more permanents are eligible under it than
	// it allows. An unbinding cap is dropped entirely, so its members
	// are not dragged into the prompt for nothing.
	binds := make([]bool, len(caps))
	for i, bc := range caps {
		n := 0
		for _, e := range own {
			if e.inCap[i] {
				n++
			}
		}
		binds[i] = n > bc.cap.Max
	}
	plan := &untapChoicePlan{}
	// The binding caps, renumbered: plan.capMax[c] is the c-th binding
	// cap and plan.members[c] is its membership over plan.candidates.
	var binding []int
	for i := range caps {
		if binds[i] {
			binding = append(binding, i)
			plan.capMax = append(plan.capMax, caps[i].cap.Max)
		}
	}
	plan.members = make([][]bool, len(binding))
	for _, e := range own {
		inQuestion := e.optional
		for _, i := range binding {
			if e.inCap[i] {
				inQuestion = true
				break
			}
		}
		if !inQuestion {
			plan.settled = append(plan.settled, e.id)
			continue
		}
		plan.candidates = append(plan.candidates, e.id)
		plan.optional = append(plan.optional, e.optional)
		for c, i := range binding {
			plan.members[c] = append(plan.members[c], e.inCap[i])
		}
	}
	if len(plan.candidates) == 0 {
		return nil
	}
	plan.settled = append(settled, plan.settled...)
	return plan
}

// queueUntapChoiceLocked asks the active player CR 502.3's question and
// stashes the rest of the untap step behind the answer.
//
// The prompt is a ChooseCardsPrompt with a different kind stamped on
// it: one shape, one payload, one validation function, one wire
// projection and one picker, shared rather than copied (ADR 0070
// Decision 2).
//
// Caller must hold g.mu in write mode.
func (g *Game) queueUntapChoiceLocked(activePlayer uuid.UUID, plan *untapChoicePlan) {
	min, max := plan.bounds()
	settled := append([]uuid.UUID(nil), plan.settled...)
	id := g.QueueChooseCardsForEffect(ChooseCardsPrompt{
		Chooser:  activePlayer,
		Question: "Untap step — choose which permanents untap",
		Cards:    append([]uuid.UUID(nil), plan.candidates...),
		Min:      min,
		Max:      max,
		// The battlefield re-check every asynchronous card pick gets:
		// a permanent chosen here can be gone by the time the answer
		// arrives (nothing can respond during the untap step, but an
		// admin move or a concession can happen while the prompt is
		// open).
		Zone: ZoneBattlefield,
		Validate: func(picked []Card) bool {
			set := make(map[uuid.UUID]bool, len(picked))
			for _, c := range picked {
				set[c.InstanceID] = true
			}
			return plan.legal(set)
		},
		Then: func(g *Game, picks []uuid.UUID) error {
			g.finishUntapStepLocked(append(settled, picks...))
			return nil
		},
	})
	// The kind is stamped after the queue rather than threaded through
	// ChooseCardsPrompt: the prompt struct describes the QUESTION, and
	// every field of it is the same for both kinds.
	if _, c := g.findChoiceLocked(id); c != nil {
		c.Kind = PendingChoiceUntapChoice
	}
}

// finishUntapStepLocked untaps the determined set and ends the step.
// Called only from the prompt's continuation; the unpaused path does
// the same two things inline in performUntapStepLocked.
//
// `ids` is the WHOLE set — the permanents that were never in question
// and the ones the player picked — so the untap is one batch, which is
// what CR 502.3's "then they untap them all simultaneously" says.
//
// A stale answer is tolerated rather than trusted, the tolerance #701
// and #725 each had to add to one resume path: if the cursor has been
// walked past the prompt by hand, the player's answer is still
// honoured but the step is not exited a second time.
//
// Caller must hold g.mu in write mode.
func (g *Game) finishUntapStepLocked(ids []uuid.UUID) {
	for _, id := range ids {
		g.untapPermanentByIDLocked(id)
	}
	if g.State != StateActive || g.Turn.Step != StepUntap || g.MulligansOpen {
		return
	}
	g.exitUntapStepLocked()
}

// exitUntapStepLocked ends the untap step and moves the cursor on.
// Untap grants no priority (CR 502.4), so "exit" means the next step's
// entry hook, immediately, in this same write lock.
//
// Two callers, which are the two ways the untap step's turn-based
// actions can finish: the StepUntap case of the step-entry hook
// (game.go), and the untap-choice prompt's continuation above. One
// function, for exitCleanupStepLocked's reason — those two sites
// deciding separately is how #661's discard path inherited a bug.
//
// Caller must hold g.mu in write mode.
func (g *Game) exitUntapStepLocked() {
	g.advanceCursorLocked()
	g.runStepEntryHooksLocked()
}

// ResolveUntapChoice answers a PendingChoiceUntapChoice: `picks` are
// the permanents the active player chose to untap.
//
// It is ResolveChooseCards with a different kind check — the prompt
// carries a choose-cards payload and a choose-cards continuation, and
// the validation is the same function. Caller must NOT hold g.mu.
func (g *Game) ResolveUntapChoice(choiceID, chooserID uuid.UUID, picks []uuid.UUID) error {
	return g.resolveCardSetPick(PendingChoiceUntapChoice, choiceID, chooserID, picks)
}
