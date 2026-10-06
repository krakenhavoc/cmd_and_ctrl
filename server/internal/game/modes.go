package game

import (
	"fmt"
	"slices"
	"sort"
	"sync"

	"github.com/google/uuid"
)

// modes.go — modal spells, triggers and activated abilities
// (CR 700.2). A card with "Choose one —" / "Choose two —" text
// declares a ModeSpec: the option labels, how many must be chosen,
// whether the same one may be chosen more than once, and — per
// option — the target clause list that governs the announcement when
// that option is picked.
//
// ONE ModeSpec, THREE owners (#764, ADR 0065 §3): Spec.Modes for a
// spell, TriggeredAbility.Modes for a trigger, and
// effects.ActivatedAbility.Modes for an activated ability. The
// difference between them is only WHEN the choice is made — CR
// 601.2b at announce for a spell, CR 603.3c as the trigger is put on
// the stack (through the mode_pick prompt), CR 602.2b at activation
// — never what is chosen or how it is validated.
//
// Modes ride CastSpellParams.Modes / ActivateAbilityParams.Modes as
// option indexes and land on StackItem.Modes, which since #764 is a
// MULTISET IN ANNOUNCE ORDER: Mystic Confluence choosing its draw
// mode three times is [0, 0, 0], and each occurrence gets its own
// target group through TargetRef.Mode. The catalog reads the choice
// back through effects.Context.HasMode / ModeCount / Modes, or lets
// the engine dispatch each chosen bullet's ModeOption.Effect in
// PRINTED order (CR 608.2c, PrintedModeOrder; the storage order is
// what pairs an occurrence with its targets, the run order is not).

// ModeOption is one bullet of a modal spell or ability.
type ModeOption struct {
	// Label is the oracle bullet, shown verbatim in the picker:
	// "Exile target player's graveyard."
	Label string

	// Targets is the target clause LIST this option adds to the
	// announcement. Nil for untargeted options. A bullet with two
	// clauses ("destroy target artifact and target enchantment") is
	// one spec with a Rest, exactly as a card-level clause list is.
	Targets *TargetSpec

	// Effect is this bullet's body, run at resolution in announce
	// order (CR 608.2c), once per OCCURRENCE — a mode chosen twice
	// under CR 700.2d runs twice. `occurrence` is the index into
	// StackItem.Modes, so the effect reads its own target group
	// (effects.Context.ModeTargets) rather than the item's whole
	// list.
	//
	// Optional. A modal SPELL may leave it nil and branch inside its
	// OnResolve on ctx.HasMode(i) instead — the older shape, still
	// supported, still reading the same data. A modal trigger or
	// activated ability has no OnResolve to branch in, so declaring
	// the bullet's body here is what lets those two be modal without
	// each card hand-writing a switch.
	//
	// Runs under g.mu held in write mode: MUST NOT call public
	// locking mutators. Added by #764.
	Effect func(g *Game, item *StackItem, occurrence int) error

	// Cost is CR 702.172a's Spree: "As an additional cost to cast
	// this spell, pay the costs associated with those modes chosen
	// this way." Brace notation ("{1}{U}"), paid IF AND ONLY IF this
	// bullet is chosen, on top of the spell's own cost and every
	// OTHER chosen bullet's. Empty for an ordinary modal bullet —
	// every modal card before S45 leaves it unset and pays nothing
	// extra for choosing it.
	//
	// Mana only, on purpose: every printed Spree card (Three Steps
	// Ahead, Explosive Derailment, Insatiable Avarice, Caught in the
	// Crossfire and the rest of the cycle) prices its bullets in mana
	// alone. A card-shaped mode cost ("discard a card" per bullet)
	// would need the enumerator's cost-payment search widened to a
	// THIRD axis beside modes and targets, and the flat discard /
	// sacrifice wire lists extended to carry a per-mode slice —
	// neither of which any catalog card asks for yet. Left as a
	// string rather than an *AdditionalCost for the same reason: a
	// mostly-empty struct enforced-empty by a boot panic is worse
	// than the honest field, and swapping the type is the whole of
	// what the future PR would do. Added by ADR 0065's 2026-09-23
	// amendment.
	Cost string
}

// ModeSpec declares a modal spell's or ability's choice.
type ModeSpec struct {
	// Prompt is the header line — "Choose one", "Choose two".
	Prompt string

	Options []ModeOption

	// Min / Max bound the number of options chosen. "Choose one" is
	// 1 / 1; "choose one or both" is 1 / 2; "choose two" is 2 / 2;
	// "choose one or more" is 1 / len(Options); "choose up to one"
	// is 0 / 1.
	Min, Max int

	// Repeatable is CR 700.2d — "you may choose the same mode more
	// than once" (Mystic Confluence, Rootcast Apprenticeship). With
	// it set the announced list is a multiset and each occurrence
	// carries its own targets; without it, a repeated index is
	// ErrInvalidParam at announce, which is what every printed modal
	// card that does not say the words expects. Added by #764.
	Repeatable bool

	// RaisedMax / RaisedMin / RaiseMaxIf are a CONDITIONAL mode count
	// (#1590, #1655). While RaiseMaxIf holds, the upper bound is
	// RaisedMax instead of Max, and the lower bound is RaisedMin
	// instead of Min when RaisedMin is set. Three printed shapes:
	//
	//   - "If you control a commander as you cast this spell, you MAY
	//     choose both instead" (Jeska's Will, Flame of Anor): the
	//     maximum rises and Min stays, because one bullet is still a
	//     legal answer. OrUpToIf.
	//   - "If it was kicked, choose both instead" (Depth Defiler), "If
	//     there are four or more card types among cards in your
	//     graveyard, choose both instead" (Prophetic Titan): no "may",
	//     so the MINIMUM rises with it and one bullet is no longer an
	//     answer. InsteadIf, RaisedMin == RaisedMax.
	//   - "If this spell was kicked, choose any number instead"
	//     (Inscription of Ruin): the maximum becomes every bullet and
	//     Min stays. AnyNumberIf.
	//
	// Read through modeBoundsLocked / ModeBoundsForEffect and nowhere
	// else, at the moment the choice is MADE: CR 601.2b's announce for
	// a spell, CR 602.2b's activation, CR 603.3c's mode_pick for a
	// trigger. The answer is recorded in StackItem.Modes and is never
	// re-asked, so a commander that leaves after the announcement
	// changes nothing — "as you cast this spell" is a check, not a
	// duration. The view stamps the same answer per caster onto
	// ModeSpecView.Min / Max, and the bot's enumerator bounds its
	// selections by it, so the picker, the enumerator and the gate
	// cannot disagree.
	//
	// RaiseMaxIf is required with either raise and refused without
	// one; effects.Register panics on a RaisedMax that does not exceed
	// Max, and on a RaisedMin outside Min..RaisedMax. Added by ADR
	// 0065's 2026-09-27 amendment; RaisedMin by its 2026-09-28 one.
	//
	// RaiseMaxIf is a KEY, not a func (ADR 0041 phase 3, #1497): a
	// ModeSpec is reachable from Game through a trigger's paused
	// mode_pick frame, and the closure ratchet
	// (testdata/closure_fields.txt) allows no new func-typed route. The
	// predicate lives in a registry beside this file, registered once
	// at init by ModeCondition / ModeConditionOnAnnouncement, and the
	// spec carries only its name. The field keeps its #1590 name
	// although it now governs the minimum too.
	RaisedMax  int
	RaisedMin  int
	RaiseMaxIf ModeCountCondition

	// NotChosen is "choose one that hasn't been chosen [this turn]"
	// (ADR 0097, #1749): the options this OBJECT's ability has already
	// chosen are withheld from its later choices, for the rest of the
	// turn (ModeMemoryThisTurn) or for as long as the object exists
	// (ModeMemoryEver). The zero value is no restriction.
	//
	// Read through choosableModeOptionsLocked, the one filter every
	// mode path asks, with the ability's identity (ModeAbility); a
	// caller with no identity — a spell — excludes nothing.
	// effects.Register refuses it on a Repeatable spec (CR 700.2d
	// says the opposite) and on a spell's Spec.Modes (no spell prints
	// it, and a spell has no object to remember with). See
	// mode_memory.go.
	NotChosen ModeMemory

	// Escalate is CR 702.120a's "Escalate [cost]": "For each mode you
	// choose beyond the first as you cast this spell, you pay an
	// additional [cost]." Paid (len(modes) - 1) times, in the cast's
	// CR 601.2f cost plan beside the card's own additional cost, so
	// its mana joins the total (AddModeCostMana), its discards ride
	// discard_ids and its taps ride teamwork_ids, all validated by the
	// ONE validator and refused when the caster cannot pay (#2126).
	// Nil for every other modal card. Build it with effects.Escalate /
	// EscalateDiscard / EscalateTapCreature.
	Escalate *EscalateCost
}

// EscalateCost is the price of CR 702.120a's escalate: what ONE mode
// beyond the first costs. A plain struct of numbers and a string, not an
// AdditionalCost, on purpose: a ModeSpec is reachable from a stack
// item, and AdditionalCost's target-clause fields would add closure
// routes to the restore-point census for components an escalate cost
// can never carry.
type EscalateCost struct {
	// ManaCost is the mana owed per extra mode, brace notation.
	ManaCost string
	// DiscardCards is the cards discarded per extra mode.
	DiscardCards int
	// TapCreatures is the untapped creatures the caster controls that
	// are tapped per extra mode (Collective Effort), any power.
	TapCreatures int
	// Label is the clause as printed ("Escalate {G}").
	Label string
}

// Empty reports whether the cost demands nothing. Nil-safe.
func (e *EscalateCost) Empty() bool {
	return e == nil || (e.ManaCost == "" && e.DiscardCards == 0 && e.TapCreatures == 0)
}

// additional is the cost as one entry of the cast's payment plan.
func (e *EscalateCost) additional() AdditionalCost {
	return AdditionalCost{ManaCost: e.ManaCost, DiscardCards: e.DiscardCards, TapCreatures: e.TapCreatures, Label: e.Label}
}

// EscalateExtra is how many times the escalate cost is owed for the
// announced modes: one per mode beyond the first (CR 702.120a). Zero
// for a spec without escalate and for one mode or none.
func EscalateExtra(ms *ModeSpec, modes []int) int {
	if ms == nil || ms.Escalate == nil || len(modes) < 2 {
		return 0
	}
	return len(modes) - 1
}

// escalatePayments is the plan entries the escalate cost adds to one
// cast: EscalateExtra copies of the cost, index -2 so they fold into
// neither the optional-cost record nor the mandatory slot.
func escalatePayments(ms *ModeSpec, modes []int) []costPayment {
	n := EscalateExtra(ms, modes)
	if n == 0 {
		return nil
	}
	out := make([]costPayment, n)
	for i := range out {
		out[i] = costPayment{cost: ms.Escalate.additional(), index: -2}
	}
	return out
}

// ModeCountCondition names a registered "if <condition> as you cast
// this spell" predicate for a conditional mode count (#1590). The key
// is unexported, so the only way to hold a non-zero one is
// ModeCondition / ModeConditionOnAnnouncement, which is what makes a
// func literal on a ModeSpec a compile error.
type ModeCountCondition struct{ key string }

// Key is the condition's registry key.
func (c ModeCountCondition) Key() string { return c.key }

// IsZero reports whether no condition is named.
func (c ModeCountCondition) IsZero() bool { return c.key == "" }

// ModeCountQuery is what a conditional mode count's predicate is
// asked about (#1655): who is choosing, and what was announced WITH
// the choice. CR 601.2b announces the modes and the optional
// additional costs in one step — "announce … whether they will pay
// optional costs" — so "if this spell was kicked, choose any number
// instead" is answerable at the same moment as "if you control a
// commander".
type ModeCountQuery struct {
	// Chooser is the player choosing the modes — "you".
	Chooser uuid.UUID

	// OracleID is the catalog key of the card whose OptionalCosts
	// slice is the index space OptionalCosts names. Empty when no
	// optional cost can have been announced.
	OracleID string

	// OptionalCosts are the optional additional costs announced for
	// the object whose modes are being chosen, as positions in the
	// card's OptionalCosts (ADR 0073): CastSpellParams.OptionalCosts
	// for a spell being cast, the spell's PaidCost.OptionalCosts for
	// a "when you cast this spell" trigger (Depth Defiler), and the
	// permanent's CastProvenance for an ability of a permanent that
	// entered kicked.
	OptionalCosts []int
}

// OptionalCostTimes counts how many times the optional cost keyed
// `key` was announced — OptionalCostTimesPaid without needing a Card.
func (q ModeCountQuery) OptionalCostTimes(key string) int {
	return optionalCostTimesFor(q.OracleID, q.OptionalCosts, key)
}

// Kicked is CR 702.33's "if it was kicked": kicker or multikicker
// announced at least once.
func (q ModeCountQuery) Kicked() bool {
	return q.OptionalCostTimes(KickerKey)+q.OptionalCostTimes(MultikickerKey) > 0
}

var modeCountConditions = struct {
	sync.RWMutex
	byKey map[string]func(g *Game, q ModeCountQuery) bool
}{byKey: map[string]func(g *Game, q ModeCountQuery) bool{}}

// ModeCondition registers a conditional-mode-count predicate over the
// board under `key` and returns its name — "if you control a
// commander". Call it once, from a package-level var in the file that
// uses it. Panics on an empty key, a nil predicate or a duplicate,
// each of which is a card-file bug that would otherwise ship a mode
// count that never rises.
//
// The predicate is read-only and runs under g.mu, so it must not call
// a public locking accessor; `chooser` is the player choosing the
// modes — "you" in "if you control a commander".
func ModeCondition(key string, fn func(g *Game, chooser uuid.UUID) bool) ModeCountCondition {
	if fn == nil {
		return ModeConditionOnAnnouncement(key, nil)
	}
	return ModeConditionOnAnnouncement(key, func(g *Game, q ModeCountQuery) bool { return fn(g, q.Chooser) })
}

// ModeConditionOnAnnouncement registers a predicate that may also read
// what was announced with the choice — "if this spell was kicked"
// (#1655). Same contract and same panics as ModeCondition.
func ModeConditionOnAnnouncement(key string, fn func(g *Game, q ModeCountQuery) bool) ModeCountCondition {
	if key == "" {
		panic("game: ModeCondition with an empty key")
	}
	if fn == nil {
		panic(fmt.Sprintf("game: mode condition %q has no predicate", key))
	}
	modeCountConditions.Lock()
	defer modeCountConditions.Unlock()
	if _, dup := modeCountConditions.byKey[key]; dup {
		panic(fmt.Sprintf("game: mode condition %q registered twice", key))
	}
	modeCountConditions.byKey[key] = fn
	return ModeCountCondition{key: key}
}

// holds evaluates the named condition. An unknown or zero key holds
// for nobody — the printed bound, the weaker reading.
func (c ModeCountCondition) holds(g *Game, q ModeCountQuery) bool {
	if c.key == "" {
		return false
	}
	modeCountConditions.RLock()
	fn, ok := modeCountConditions.byKey[c.key]
	modeCountConditions.RUnlock()
	return ok && fn(g, q)
}

// OrUpToIf declares the "you MAY choose N instead" count on a spec
// built by one of the catalog constructors — "choose one; if you
// control a commander as you cast this spell, you may choose both
// instead" is ChooseOne(…).OrUpToIf(2, effects.YouControlACommander).
// Mutates and returns the spec so a card file reads like its oracle
// text, the way TargetSpec.WithCount does.
func (ms *ModeSpec) OrUpToIf(n int, cond ModeCountCondition) *ModeSpec {
	ms.RaisedMax, ms.RaiseMaxIf = n, cond
	return ms
}

// InsteadIf declares the "choose N instead" count with no "may" —
// "If it was kicked, choose both instead" is
// ChooseOne(…).InsteadIf(2, effects.WasKicked). The minimum rises
// with the maximum, so one bullet is no longer an answer (#1655).
func (ms *ModeSpec) InsteadIf(n int, cond ModeCountCondition) *ModeSpec {
	ms.RaisedMax, ms.RaisedMin, ms.RaiseMaxIf = n, n, cond
	return ms
}

// AnyNumberIf declares "choose any number instead" / "choose one or
// more instead": the maximum becomes every printed bullet and the
// minimum stays (#1655).
func (ms *ModeSpec) AnyNumberIf(cond ModeCountCondition) *ModeSpec {
	ms.RaisedMax, ms.RaiseMaxIf = len(ms.Options), cond
	return ms
}

// modeBoundsLocked is the lower and upper bound on how many options
// may be picked from `ms` for the choice `q` describes: the raised
// bounds while the spec's condition holds, Min / Max otherwise
// (#1590, #1655). A hi of 0 still means unbounded, as Max always has.
//
// Caller must hold g.mu.
func (g *Game) modeBoundsLocked(ms *ModeSpec, q ModeCountQuery) (lo, hi int) {
	if ms == nil {
		return 0, 0
	}
	if ms.RaisedMax > ms.Max && ms.RaiseMaxIf.holds(g, q) {
		return max(ms.Min, ms.RaisedMin), ms.RaisedMax
	}
	return ms.Min, ms.Max
}

// ModeBoundsForEffect is modeBoundsLocked on the *ForEffect surface,
// for the callers already under g.mu that must quote the same bounds
// the announce gate enforces — the bot's move enumerator and the
// protocol projection (#1590, #1655).
func (g *Game) ModeBoundsForEffect(ms *ModeSpec, q ModeCountQuery) (lo, hi int) {
	return g.modeBoundsLocked(ms, q)
}

// modeQueryForSourceLocked is the ModeCountQuery for a choice made on
// behalf of `source` — a trigger's mode_pick or an activated ability.
// The optional costs are the source's own announcement: the spell's
// PaidCost while it is on the stack (a "when you cast this spell"
// trigger — Depth Defiler), the permanent's CastProvenance once it has
// entered (#1655).
//
// Caller must hold g.mu.
func (g *Game) modeQueryForSourceLocked(source Card, chooser uuid.UUID) ModeCountQuery {
	q := ModeCountQuery{Chooser: chooser, OracleID: CatalogKey(source)}
	if item, ok := g.StackMeta[source.InstanceID]; ok && item != nil && item.Kind == StackItemSpell {
		q.OptionalCosts = item.Paid.OptionalCosts
		return q
	}
	q.OptionalCosts = source.Provenance.OptionalCosts
	return q
}

// ModeQueryForSourceForEffect is modeQueryForSourceLocked on the
// *ForEffect surface, for the enumerator's and the view's
// activated-ability rows, which must ask the question the activation
// gate asks.
func (g *Game) ModeQueryForSourceForEffect(source Card, chooser uuid.UUID) ModeCountQuery {
	return g.modeQueryForSourceLocked(source, chooser)
}

// CatalogModeSpec is the catalog hook the effects package wires at
// init. Nil, or a nil return, means the card isn't modal.
var CatalogModeSpec func(oracleID string) *ModeSpec

// ModeSpecFor returns the structured modes for a card, or nil.
func ModeSpecFor(oracleID string) *ModeSpec {
	if CatalogModeSpec == nil || oracleID == "" {
		return nil
	}
	return CatalogModeSpec(oracleID)
}

// validateModes is the announce-time gate for a modal announcement:
// every index must name an option, none may repeat unless the spec
// is Repeatable (CR 700.2d), and the count must fall within
// lo..hi, the bounds modeBoundsLocked quoted for this announcement
// (#1590, #1655 — the spec's own Min / Max unless a conditional mode
// count raised them). A nil spec (non-catalog card, or a catalog card that
// isn't modal) keeps the S13.1 free-form behaviour: any indexes the
// client sends are recorded for opponents to see and players resolve
// by hand.
func validateModes(spec *ModeSpec, lo, hi int, modes []int) error {
	if spec == nil {
		return nil
	}
	seen := make(map[int]bool, len(modes))
	for _, m := range modes {
		if m < 0 || m >= len(spec.Options) {
			return ErrInvalidParam
		}
		if seen[m] && !spec.Repeatable {
			return ErrInvalidParam
		}
		seen[m] = true
	}
	if len(modes) < lo || (hi > 0 && len(modes) > hi) {
		return ErrInvalidParam
	}
	return nil
}

// ModeCount is how many times option i was chosen (CR 700.2d).
func ModeCount(modes []int, i int) int {
	n := 0
	for _, m := range modes {
		if m == i {
			n++
		}
	}
	return n
}

// castClauseSources is the (card-level spec, mode spec) pair that
// governs a cast of this card. Exactly one is non-nil for a catalog
// card that targets: effects.Register refuses a card declaring both,
// and the modal path derives its clause list per chosen occurrence
// through AnnouncedClauses.
func castClauseSources(oracleID string) (*TargetSpec, *ModeSpec) {
	if spec := TargetSpecFor(oracleID); spec != nil {
		return spec, nil
	}
	return nil, ModeSpecFor(oracleID)
}

// castTargetSpecForItem is the clause list a SPELL item on the stack
// was announced under, for the callers that still want one spec
// rather than the step list: the CR 707.10 copy re-target (which
// re-targets the first clause only — ADR 0065 "Out of scope") and
// the S22 alternative-cost rewrite.
//
// For a modal item it returns the first chosen option that targets,
// which is what the copy path did before per-mode targets existed.
func castTargetSpecForItem(oracleID string, item *StackItem) *TargetSpec {
	if item == nil {
		return TargetSpecFor(oracleID)
	}
	spec, ms := castClauseSources(oracleID)
	if spec == nil && ms != nil {
		for _, m := range item.Modes {
			if m >= 0 && m < len(ms.Options) && ms.Options[m].Targets != nil {
				spec = ms.Options[m].Targets
				break
			}
		}
	}
	spec = TargetSpecUnderAlternativeCost(spec, AlternativeCostByKey(oracleID, item.AltCost))
	// ADR 0089 §3: the same rewrite the announce path applied, so a
	// restored or copied promised Long River's Pull is still judged
	// under "target spell".
	return TargetSpecUnderOptionalCosts(spec, OptionalCostsFor(oracleID), item.Paid.OptionalCosts)
}

// PrintedModeOrder is the order a modal spell or ability's chosen
// modes are carried out: the occurrence indexes of `modes`, sorted by
// option (the order the modes are WRITTEN on the card) with repeats of
// one option keeping their announce order (CR 608.2c, CR 700.2d). The
// stored StackItem.Modes stays in announce order, because an
// occurrence's index is what pairs it with its target group
// (TargetRef.Mode); only the order the bodies RUN in is printed order.
// The owner decided this on 2026-09-30 (#1653, ADR 0065 §3 amendment).
func PrintedModeOrder(modes []int) []int {
	occs := make([]int, len(modes))
	for i := range occs {
		occs[i] = i
	}
	sort.SliceStable(occs, func(a, b int) bool { return modes[occs[a]] < modes[occs[b]] })
	return occs
}

// runChosenModeEffectsLocked runs each chosen bullet's ModeOption
// Effect in PRINTED order, once per occurrence (CR 608.2c, and CR
// 700.2d for a repeated mode), whatever order the modes were announced
// in. A nil Effect means the card resolves its modes inside its own
// OnResolve instead, which is the older and still-supported shape.
//
// Caller must hold g.mu in write mode.
func (g *Game) runChosenModeEffectsLocked(item *StackItem, ms *ModeSpec) {
	if item == nil || ms == nil {
		return
	}
	for _, occ := range PrintedModeOrder(item.Modes) {
		opt := item.Modes[occ]
		if opt < 0 || opt >= len(ms.Options) {
			continue
		}
		fn := ms.Options[opt].Effect
		if fn == nil {
			continue
		}
		if err := fn(g, item, occ); err != nil {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				Actor:    item.Controller,
				Source:   item.SourceCardID,
				ErrorMsg: err.Error(),
			})
		}
	}
}

// choosableModeOptionsLocked lists the option indexes a chooser may
// pick right now: every option, minus those the ability `ab` has
// already chosen under the spec's NotChosen restriction (ADR 0097),
// minus those whose clause list cannot be filled from the current
// board (CR 603.3d — an option with no legal target is not on offer).
// `src` names the spell or ability doing the choosing, because whether
// a clause is fillable depends on the source under CR 702.16b as well
// as on the chooser (#662). `ab` is the zero ModeAbility for a spell,
// which excludes nothing.
//
// The used modes are dropped FIRST, so the harvest check, the
// mode_pick prompt, the activation gate, the enumerator and the view
// all see an exhausted ability the same way.
//
// Caller must hold g.mu.
func (g *Game) choosableModeOptionsLocked(src TargetSource, ms *ModeSpec, ab ModeAbility) []int {
	if ms == nil {
		return nil
	}
	used := g.modesChosenLocked(ms, ab)
	out := make([]int, 0, len(ms.Options))
	for i, o := range ms.Options {
		if slices.Contains(used, i) {
			continue
		}
		if o.Targets != nil && g.anyClauseUnfillableLocked(src, AnnouncedClauses(o.Targets, nil, nil)) {
			continue
		}
		out = append(out, i)
	}
	return out
}

// ModeLabels is the oracle bullet of each chosen mode of this item,
// in the order they will be carried out (printed order, CR 608.2c) and
// with repeats — what the stack overlay shows instead of a row of
// indexes. Empty for a non-modal item and for
// one whose ModeSpec could not be re-derived after a restore.
func (s *StackItem) ModeLabels() []string {
	if s == nil || s.modeSpec == nil || len(s.Modes) == 0 {
		return nil
	}
	out := make([]string, 0, len(s.Modes))
	for _, occ := range PrintedModeOrder(s.Modes) {
		m := s.Modes[occ]
		if m < 0 || m >= len(s.modeSpec.Options) {
			continue
		}
		out = append(out, s.modeSpec.Options[m].Label)
	}
	return out
}

// ChoosableModeOptionsForEffect is choosableModeOptionsLocked on the
// *ForEffect surface: which of a ModeSpec's options the spell or
// ability `src` could take right now, for callers already under g.mu
// — the bot's move enumerator and the protocol projection. `ab` names
// the ability whose "hasn't been chosen" memory applies; the zero
// value (a spell) excludes nothing.
func (g *Game) ChoosableModeOptionsForEffect(src TargetSource, ms *ModeSpec, ab ModeAbility) []int {
	return g.choosableModeOptionsLocked(src, ms, ab)
}

// EnoughChoosableModes reports whether `n` takeable options can fill a
// selection of Min. A Repeatable spec (CR 700.2d) needs only ONE — it
// may take the same bullet Min times, which is the whole point of
// Mystic Confluence's "choose three, you may choose the same mode
// more than once" with two of its bullets unfillable.
func EnoughChoosableModes(n int, ms *ModeSpec) bool {
	if ms == nil || n == 0 {
		return false
	}
	if ms.Repeatable {
		return true
	}
	return n >= ms.Min
}

// EscalatePayableExtraForEffect is the most modes beyond the first
// `playerID` could pay escalate's non-mana half for right now (CR
// 702.120a): the cards in hand other than the spell itself divided by
// the discards per mode, the untapped creatures divided by the taps per
// mode, whichever is smaller, capped at one fewer than the spell's most
// modes. Mana is deliberately not asked — CR 601.2g lets the caster
// activate mana abilities after announcing, the posture
// AdditionalCostBranchPayableLocked takes. -1 for a spec without
// escalate.
//
// ONE function, read by the view's `modes.escalate.max_extra` and
// matched by what the bot enumerator and CastSpell enforce, so no mode
// count is offered that cannot be paid (#2126).
//
// Caller must hold g.mu (read or write).
func (g *Game) EscalatePayableExtraForEffect(playerID, castID uuid.UUID, ms *ModeSpec) int {
	if ms == nil || ms.Escalate == nil {
		return -1
	}
	best := max(min(ms.Max, len(ms.Options))-1, 0)
	e := ms.Escalate
	if e.DiscardCards > 0 {
		p := g.playerByIDLocked(playerID)
		if p == nil {
			return 0
		}
		n := 0
		for _, c := range p.Hand.Cards {
			if c.InstanceID != castID {
				n++
			}
		}
		best = min(best, n/e.DiscardCards)
	}
	if e.TapCreatures > 0 {
		best = min(best, len(g.TapCreaturesOptionsForEffect(playerID))/e.TapCreatures)
	}
	return best
}

// AddModeCostMana is CR 702.172a's Spree, the mana half: the sum of
// every CHOSEN mode's own Cost, added into `cost` at CR 601.2f beside
// AdditionalCostMana (ADR 0073 §3) — the same point in the
// precedence and the same reason. Thalia taxes a Spree spell once for
// the whole announced total, and the cast path, the bot enumerator
// and the auto-tap preview must price one mode selection identically
// (#544), so there is exactly one walk of this arithmetic.
//
// A nil ModeSpec, or one where no chosen option carries a Cost,
// changes nothing — every modal card that predates S45 reprices to
// the exact number it always did.
//
// `modes` is a MULTISET in announce order (CR 700.2d): a repeated
// option pays its Cost once per occurrence, which is correct by
// construction since the loop walks the multiset rather than a set of
// distinct indices.
func AddModeCostMana(cost ParsedCost, ms *ModeSpec, modes []int) (ParsedCost, error) {
	if ms == nil {
		return cost, nil
	}
	// CR 702.120a: escalate's mana, once per mode beyond the first.
	if n := EscalateExtra(ms, modes); n > 0 && ms.Escalate.ManaCost != "" {
		add, err := ParseCost(ms.Escalate.ManaCost)
		if err != nil {
			return cost, fmt.Errorf("%w for %s: %w", ErrUnparseableCost, ms.Escalate.Label, err)
		}
		for range n {
			cost.Generic += add.Generic
			cost.Required = append(cost.Required, add.Required...)
			cost.XSlots += add.XSlots
			cost.HasPhyrexian = cost.HasPhyrexian || add.HasPhyrexian
			cost.HasSnow = cost.HasSnow || add.HasSnow
		}
	}
	for _, m := range modes {
		if m < 0 || m >= len(ms.Options) || ms.Options[m].Cost == "" {
			continue
		}
		add, err := ParseCost(ms.Options[m].Cost)
		if err != nil {
			return cost, fmt.Errorf("%w for %s: %w", ErrUnparseableCost, ms.Options[m].Label, err)
		}
		cost.Generic += add.Generic
		cost.Required = append(cost.Required, add.Required...)
		cost.XSlots += add.XSlots
		cost.HasPhyrexian = cost.HasPhyrexian || add.HasPhyrexian
		cost.HasSnow = cost.HasSnow || add.HasSnow
	}
	return cost, nil
}
