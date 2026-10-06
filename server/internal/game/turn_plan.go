package game

import (
	"strings"

	"github.com/google/uuid"
)

// turn_plan.go — the rest of the turn is data (ADR 0059 Decisions 3,
// 4 and 8, sub-PR 2b, #753).
//
// The cursor used to walk turnSequence by index, so nothing could add
// a phase or a step. Now the steps still to come in THIS turn are a
// list, Game.TurnPlan, and advanceCursorLocked pops its head. A turn
// begins with the template (turnSequence after untap, phase ids 1..5),
// and an effect that adds a phase or a step splices it in:
//
//   - AddPhasesForEffect: "after this phase, there is an additional
//     combat phase" (Aurelia, Karlach, Hellkite Charger) and "after
//     this main phase, there is an additional combat phase followed by
//     an additional main phase" (Relentless Assault, CR 500.8).
//   - AddStepAfterCurrentForEffect: "there is an additional end step
//     after this step" (Y'shtola Rhul, CR 500.9).
//
// Both put the new entries directly after the anchor and AHEAD of
// anything already added after it, which is CR 500.8's "the most
// recently created phase will occur first" with no rule of its own.
//
// #717's first-strike combat damage step is an ordinary plan entry of
// every combat phase. Whether the turn HAS that step is still decided
// as the cursor lands on it (Game.stepExistsLocked): the question is
// about the board as the combat damage step would begin, which no plan
// written earlier in the turn can know. So the plan lists it and the
// entry hook walks through it when no combatant has first or double
// strike — exactly the behaviour #717 shipped, for every combat.
//
// Why the plan is on Game rather than inside Turn: Turn is copied by
// value everywhere (the wire view, event payloads, `prev := g.Turn`),
// and a slice inside it would alias between copies.

// PhaseKind is a phase FAMILY: what AddPhasesForEffect adds, and what
// Turn.PhaseOrdinal counts. Precombat and postcombat main are one
// family (CR 505.1b: "first main phase", "second main phase" count
// main phases of the turn).
type PhaseKind string

const (
	PhaseKindBeginning PhaseKind = "beginning"
	PhaseKindMain      PhaseKind = "main"
	PhaseKindCombat    PhaseKind = "combat"
	PhaseKindEnding    PhaseKind = "ending"
)

// PhaseKindOf returns the family of a phase, or "" for an unknown one.
func PhaseKindOf(p Phase) PhaseKind {
	switch p {
	case PhaseBeginning:
		return PhaseKindBeginning
	case PhasePrecombatMain, PhasePostcombatMain:
		return PhaseKindMain
	case PhaseCombat:
		return PhaseKindCombat
	case PhaseEnding:
		return PhaseKindEnding
	}
	return ""
}

// PlannedStep is one step still to come in this turn.
type PlannedStep struct {
	Step Step `json:"step"`
	// PhaseID is the phase instance the step belongs to: two combat
	// phases in one turn have two ids, so "after this phase" knows
	// where this one ends.
	PhaseID int `json:"phaseId"`
}

// templatePhaseCount is how many phase ids the template itself uses.
// The first phase an effect adds is templatePhaseCount+1.
const templatePhaseCount = 5

// templatePhaseID is the phase id a template step belongs to:
// beginning 1, precombat main 2, combat 3, postcombat main 4,
// ending 5.
func templatePhaseID(s Step) int {
	switch PhaseOf(s) {
	case PhaseBeginning:
		return 1
	case PhasePrecombatMain:
		return 2
	case PhaseCombat:
		return 3
	case PhasePostcombatMain:
		return 4
	case PhaseEnding:
		return 5
	}
	return 0
}

// phaseKindSteps is what an added phase of kind k expands to
// (ADR 0059 Decision 4). An added main phase is a POSTCOMBAT main
// phase: CR 505.1a makes only the first main phase of a turn
// precombat, so the Saga lore action and "at the beginning of your
// precombat main phase" never happen twice. An added combat carries
// #717's first-strike step like the template's does.
func phaseKindSteps(k PhaseKind) []Step {
	switch k {
	case PhaseKindBeginning:
		return []Step{StepUntap, StepUpkeep, StepDraw}
	case PhaseKindMain:
		return []Step{StepPostcombatMain}
	case PhaseKindCombat:
		return []Step{StepBeginCombat, StepDeclareAttackers, StepDeclareBlockers,
			StepFirstStrikeDamage, StepCombatDamage, StepEndCombat}
	case PhaseKindEnding:
		return []Step{StepEnd, StepCleanup}
	}
	return nil
}

// templateTailAfter is the template's steps after `s`, with template
// phase ids — the plan a turn has when nothing has been added.
func templateTailAfter(s Step) []PlannedStep {
	idx := indexOfStep(s)
	if idx < 0 {
		return nil
	}
	out := make([]PlannedStep, 0, len(turnSequence)-idx-1)
	for _, step := range turnSequence[idx+1:] {
		out = append(out, PlannedStep{Step: step, PhaseID: templatePhaseID(step)})
	}
	return out
}

// planCursorLocked is the cursor position the plan must have been
// built for to be trusted.
func (g *Game) planCursorLocked() TurnStep {
	return TurnStep{Turn: g.Turn.Seq, Step: g.Turn.Step}
}

// planIsCurrentLocked reports whether TurnPlan describes the rest of
// the turn from where the cursor is standing. It is false for a
// cursor moved by something other than the plan — a test that sets
// Turn.Step by hand, a snapshot written before the plan existed — and
// the plan is then rebuilt from the template (ensureTurnPlanLocked).
func (g *Game) planIsCurrentLocked() bool {
	return g.planAt.NamesAStep() && g.planAt == g.planCursorLocked()
}

// resetTurnPlanLocked fills the plan with the template's tail after
// the step the cursor is on, stamps that step's template phase id, and
// restarts the phase-id counter. Called as a turn begins, and to
// rebuild a plan the cursor has left behind.
//
// Caller must hold g.mu.
func (g *Game) resetTurnPlanLocked() {
	g.TurnPlan = templateTailAfter(g.Turn.Step)
	g.Turn.PhaseID = templatePhaseID(g.Turn.Step)
	g.NextPhaseID = templatePhaseCount
	g.planAt = g.planCursorLocked()
}

// ensureTurnPlanLocked rebuilds the plan from the template when it
// does not describe the cursor's position (planIsCurrentLocked). An
// ordinary game never needs it: the turn-began hook fills the plan and
// every step transition pops it. Anything that reads or edits the plan
// calls this first.
//
// Caller must hold g.mu in write mode.
func (g *Game) ensureTurnPlanLocked() {
	if g.planIsCurrentLocked() {
		return
	}
	next := g.NextPhaseID
	g.resetTurnPlanLocked()
	if next > g.NextPhaseID {
		// A plan rebuilt mid-turn keeps minting fresh ids.
		g.NextPhaseID = next
	}
}

// popTurnPlanLocked moves the cursor onto the plan's next step and
// reports whether there was one. An empty plan means the turn is over:
// the caller hands over to the rotation seam.
//
// The head is removed by reslicing. Nothing writes into the backing
// array in place — every insertion builds a fresh slice — so an undo
// clone sharing it can never see a later edit.
//
// Caller must hold g.mu.
func (g *Game) popTurnPlanLocked() bool {
	g.ensureTurnPlanLocked()
	if len(g.TurnPlan) == 0 {
		return false
	}
	head := g.TurnPlan[0]
	g.TurnPlan = g.TurnPlan[1:]
	if len(g.TurnPlan) == 0 {
		g.TurnPlan = nil
	}
	g.Turn.Step = head.Step
	g.Turn.Phase = PhaseOf(head.Step)
	g.Turn.PhaseID = head.PhaseID
	g.grantPriorityLocked(initialPriorityHolder(head.Step, g.Turn.ActiveSeat))
	g.planAt = g.planCursorLocked()
	return true
}

// noteStepBegunLocked stamps the ordinals as a step really begins
// (ADR 0059 Decision 8): after the CR 500.11 skip window and the
// mulligan hold, so a skipped step is not counted. The first step to
// begin in a new phase instance counts a phase of its family.
//
// Caller must hold g.mu.
func (g *Game) noteStepBegunLocked() {
	g.ensureTurnPlanLocked()
	t := &g.TurnTally
	if g.Turn.PhaseID != t.PhaseStarted {
		if t.PhasesBegun == nil {
			t.PhasesBegun = map[PhaseKind]int{}
		}
		kind := PhaseKindOf(g.Turn.Phase)
		t.PhasesBegun[kind]++
		t.PhaseStarted = g.Turn.PhaseID
		g.Turn.PhaseOrdinal = t.PhasesBegun[kind]
	}
	if t.StepsBegun == nil {
		t.StepsBegun = map[Step]int{}
	}
	t.StepsBegun[g.Turn.Step]++
	g.Turn.StepOrdinal = t.StepsBegun[g.Turn.Step]
	if g.Turn.Step == StepEnd && g.Turn.ActiveSeat >= 0 && g.Turn.ActiveSeat < len(g.Seats) {
		// #2385: an end step that really begins. A turn that is ended
		// never gets here, which is what "until your next end step"
		// reads to keep its window open across it.
		if p := g.Seats[g.Turn.ActiveSeat]; p != nil {
			p.EndStepTurn = p.TurnsBegun
		}
	}
}

// derivePrePlanOrdinalsLocked fills in the ordinals for a restore
// point written before the turn plan existed (ADR 0059 Decision 8,
// #753). Such a file carries no Turn.PhaseOrdinal or StepOrdinal and
// no StepsBegun, PhasesBegun or PhaseStarted, so without this the
// counts start again from zero at the restored step: a Relentless
// Assault cast in the restored second main phase adds a combat that
// then reads as the FIRST combat of the turn (PhaseOrdinal 1), and
// Karlach would add yet another.
//
// No phase or step can have been added in a turn written before the
// plan, so the turn so far is the template up to the cursor, and the
// counts are what noteStepBegunLocked would have recorded walking it.
// Two steps of the template are left out, both the way a live game
// leaves them out: #717's first-strike damage step before the cursor
// (it exists only when a combatant had first or double strike, which
// the file cannot say; usually it did not), and, while the mulligan
// window is open, the untap and upkeep steps it holds (they have not
// begun).
//
// Caller must hold g.mu in write mode.
func (g *Game) derivePrePlanOrdinalsLocked() {
	if g.MulligansOpen && (g.Turn.Step == StepUntap || g.Turn.Step == StepUpkeep) {
		return
	}
	idx := indexOfStep(g.Turn.Step)
	if idx < 0 {
		return
	}
	t := &g.TurnTally
	t.StepsBegun = map[Step]int{}
	t.PhasesBegun = map[PhaseKind]int{}
	t.PhaseStarted = 0
	for i, step := range turnSequence[:idx+1] {
		if step == StepFirstStrikeDamage && i != idx {
			continue
		}
		if id := templatePhaseID(step); id != t.PhaseStarted {
			t.PhasesBegun[PhaseKindOf(PhaseOf(step))]++
			t.PhaseStarted = id
		}
		t.StepsBegun[step]++
	}
	g.Turn.PhaseOrdinal = t.PhasesBegun[PhaseKindOf(PhaseOf(g.Turn.Step))]
	g.Turn.StepOrdinal = t.StepsBegun[g.Turn.Step]
}

// AnchorKind names which phase an added phase goes after.
type AnchorKind int

const (
	// AnchorThisPhase is "after this phase" (Aurelia, the Warleader;
	// Karlach; Hellkite Charger): whatever phase is in progress.
	AnchorThisPhase AnchorKind = iota + 1
	// AnchorThisMainPhase is "after this main phase" (Relentless
	// Assault, Seize the Day, Full Throttle). It adds nothing unless a
	// main phase is in progress: the rulings are explicit that a copy
	// resolving in combat or an opponent's upkeep creates no phases.
	AnchorThisMainPhase
)

// PhaseAnchor says where AddPhasesForEffect splices. Decision 4's
// AnchorNthMainPhase ("after the second main phase this turn", World
// at War) is not built: it ships with the first card that uses it.
type PhaseAnchor struct {
	Kind AnchorKind
}

// AddPhasesForEffect adds phases, in the printed order, directly after
// the anchor phase of the CURRENT turn (CR 500.8), whoever controls the
// source (the Relentless Assault and Seize the Day rulings: on an
// opponent's turn, that opponent gets the combat). CR 500.10a — "you
// get" an additional phase on someone else's turn adds nothing — is the
// card's to check, because only the card knows its wording; no card
// built so far says "you get".
//
// It returns the new phases' ids in the order they will occur, or nil
// when nothing was added: an AnchorThisMainPhase outside a main phase,
// no kinds, or an unknown kind. Emits one EventPhasesAdded.
//
// Safe from inside a resolution: it only edits the plan. The phases
// begin when the cursor reaches them.
//
// Caller must hold g.mu.
func (g *Game) AddPhasesForEffect(source uuid.UUID, anchor PhaseAnchor, kinds ...PhaseKind) []int {
	if len(kinds) == 0 {
		return nil
	}
	switch anchor.Kind {
	case AnchorThisPhase:
	case AnchorThisMainPhase:
		if PhaseKindOf(g.Turn.Phase) != PhaseKindMain {
			return nil
		}
	default:
		return nil
	}
	for _, k := range kinds {
		if phaseKindSteps(k) == nil {
			return nil
		}
	}
	g.ensureTurnPlanLocked()
	anchorID := g.Turn.PhaseID
	at := 0
	for at < len(g.TurnPlan) && g.TurnPlan[at].PhaseID == anchorID {
		at++
	}
	var added []PlannedStep
	ids := make([]int, 0, len(kinds))
	names := make([]string, 0, len(kinds))
	for _, k := range kinds {
		g.NextPhaseID++
		id := g.NextPhaseID
		ids = append(ids, id)
		names = append(names, string(k))
		for _, s := range phaseKindSteps(k) {
			added = append(added, PlannedStep{Step: s, PhaseID: id})
		}
	}
	g.TurnPlan = splicePlan(g.TurnPlan, at, added)
	g.EmitEvent(Event{
		Kind:   EventPhasesAdded,
		Actor:  g.activePlayerIDLocked(),
		Source: source,
		Amount: len(ids),
		Label:  strings.Join(names, ","),
	})
	return ids
}

// AddStepAfterCurrentForEffect adds one step directly after the step
// in progress, in the same phase (CR 500.9): Y'shtola Rhul's "there is
// an additional end step after this step". Two added the same way run
// newest first, like phases. Returns false, adding nothing, for a step
// that does not belong to the phase in progress.
//
// Emits EventPhasesAdded with Step set, so the log can tell a step
// from a phase.
//
// Caller must hold g.mu.
func (g *Game) AddStepAfterCurrentForEffect(source uuid.UUID, step Step) bool {
	if PhaseKindOf(PhaseOf(step)) == "" || PhaseKindOf(PhaseOf(step)) != PhaseKindOf(g.Turn.Phase) {
		return false
	}
	g.ensureTurnPlanLocked()
	g.TurnPlan = splicePlan(g.TurnPlan, 0, []PlannedStep{{Step: step, PhaseID: g.Turn.PhaseID}})
	g.EmitEvent(Event{
		Kind:   EventPhasesAdded,
		Actor:  g.activePlayerIDLocked(),
		Source: source,
		Amount: 1,
		Step:   step,
		Label:  string(step),
	})
	return true
}

// splicePlan returns a FRESH slice with `added` inserted at index at.
// Never an in-place insert: RestoreFrom hands the live game the backing
// array an undo snapshot holds.
func splicePlan(plan []PlannedStep, at int, added []PlannedStep) []PlannedStep {
	out := make([]PlannedStep, 0, len(plan)+len(added))
	out = append(out, plan[:at]...)
	out = append(out, added...)
	out = append(out, plan[at:]...)
	return out
}

// UpcomingStepsForEffect returns the steps still to come in this turn,
// in order — TurnView.upcoming. Read-only and safe under the read
// lock: a plan the cursor has left behind is answered from the
// template rather than rebuilt. Returns a fresh slice.
//
// Caller must hold g.mu (read or write).
func (g *Game) UpcomingStepsForEffect() []PlannedStep {
	if g.planIsCurrentLocked() {
		return clonePlan(g.TurnPlan)
	}
	return templateTailAfter(g.Turn.Step)
}

// IsFirstCombatPhaseForEffect is "if it's the first combat phase of the
// turn" (Karlach, Genji Glove, Raiyuu): a combat phase is in progress
// and it is the first of the turn. A cursor whose phase has not begun
// yet (ordinal 0) is the first.
//
// Caller must hold g.mu.
func (g *Game) IsFirstCombatPhaseForEffect() bool {
	return g.Turn.Phase == PhaseCombat && g.Turn.PhaseOrdinal <= 1
}

// IsFirstStepOfItsKindForEffect is "if it's the first <step> of the
// turn" (Y'shtola Rhul's "first end step"): the step in progress has
// not begun before this turn.
//
// Caller must hold g.mu.
func (g *Game) IsFirstStepOfItsKindForEffect() bool {
	return g.Turn.StepOrdinal <= 1
}

// clonePlan copies a plan; nil stays nil.
func clonePlan(p []PlannedStep) []PlannedStep {
	if len(p) == 0 {
		return nil
	}
	return append([]PlannedStep(nil), p...)
}
