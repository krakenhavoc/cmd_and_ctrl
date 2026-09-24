package game

// Phase is one of MTG's five high-level turn phases.
type Phase string

const (
	PhaseBeginning      Phase = "beginning"
	PhasePrecombatMain  Phase = "precombat_main"
	PhaseCombat         Phase = "combat"
	PhasePostcombatMain Phase = "postcombat_main"
	PhaseEnding         Phase = "ending"
)

// Step is one of MTG's per-turn steps. Every step belongs to exactly
// one phase; the phase field on Turn is derivable from the step via
// PhaseOf below, but stored for convenience and wire-format clarity.
//
// The sequence below is the complete MTG turn structure. Every step
// is in it, but not every step happens in every turn: a step the turn
// does not have is SKIPPED, which the cursor models by walking
// straight through it (Game.stepExistsLocked, game.go). Today the
// first-strike combat damage step is the only one of those — CR 506.1
// gives a combat two combat damage steps only when a combatant has
// first or double strike as the first one would begin.
type Step string

const (
	StepUntap            Step = "untap"
	StepUpkeep           Step = "upkeep"
	StepDraw             Step = "draw"
	StepPrecombatMain    Step = "precombat_main"
	StepBeginCombat      Step = "begin_combat"
	StepDeclareAttackers Step = "declare_attackers"
	StepDeclareBlockers  Step = "declare_blockers"

	// StepFirstStrikeDamage is the FIRST of the two combat damage
	// steps CR 510.4 gives a combat in which any attacking or
	// blocking creature has first strike or double strike as the
	// combat damage step begins. It is a step like any other: its
	// turn-based action is the first-strike damage pass, the triggers
	// that damage causes go on the stack, and the active player then
	// receives priority (CR 510.3) — which is the window that makes
	// "respond after first strike" possible at all (#717).
	//
	// When no combatant has either keyword the step does not exist
	// and the cursor walks through it without entering it, so an
	// ordinary combat is one combat damage step exactly as before.
	StepFirstStrikeDamage Step = "first_strike_damage"

	// StepCombatDamage is the regular combat damage step — the
	// SECOND one when StepFirstStrikeDamage happened, and the only
	// one when it did not. Its name and wire value are unchanged:
	// every combat still has this step.
	StepCombatDamage   Step = "combat_damage"
	StepEndCombat      Step = "end_combat"
	StepPostcombatMain Step = "postcombat_main"
	StepEnd            Step = "end"
	StepCleanup        Step = "cleanup"
)

// turnSequence is the canonical order in which steps occur within a
// single player's turn. The package is the single source of truth for
// turn structure; tests assert against this slice directly.
var turnSequence = []Step{
	StepUntap,
	StepUpkeep,
	StepDraw,
	StepPrecombatMain,
	StepBeginCombat,
	StepDeclareAttackers,
	StepDeclareBlockers,
	StepFirstStrikeDamage,
	StepCombatDamage,
	StepEndCombat,
	StepPostcombatMain,
	StepEnd,
	StepCleanup,
}

// TurnSequence returns a copy of the canonical step sequence. Callers
// may freely mutate the returned slice without affecting the package.
func TurnSequence() []Step {
	out := make([]Step, len(turnSequence))
	copy(out, turnSequence)
	return out
}

// stepPhase is the (step → phase) lookup built once at init time.
var stepPhase = map[Step]Phase{
	StepUntap:             PhaseBeginning,
	StepUpkeep:            PhaseBeginning,
	StepDraw:              PhaseBeginning,
	StepPrecombatMain:     PhasePrecombatMain,
	StepBeginCombat:       PhaseCombat,
	StepDeclareAttackers:  PhaseCombat,
	StepDeclareBlockers:   PhaseCombat,
	StepFirstStrikeDamage: PhaseCombat,
	StepCombatDamage:      PhaseCombat,
	StepEndCombat:         PhaseCombat,
	StepPostcombatMain:    PhasePostcombatMain,
	StepEnd:               PhaseEnding,
	StepCleanup:           PhaseEnding,
}

// PhaseOf returns the phase that contains a given step.
func PhaseOf(s Step) Phase {
	return stepPhase[s]
}

// NoPriority is the sentinel PriorityHolder value meaning "no player
// holds priority right now." Per CR 502.4 / 514.3, the Untap and
// Cleanup steps do not grant priority. PassPriority returns
// ErrNoPriority while PriorityHolder == NoPriority; the auto turn-
// based actions in runStepEntryHooksLocked drive the cursor through
// these steps without requiring a player click. Added in S13.
const NoPriority = -1

// Turn is the cursor into the game's turn/phase/step state machine.
// Seq identifies one turn and advances for every turn that begins.
// Round is the table-facing rotation count shared by each seat's turn
// in a round. PriorityHolder tracks which seat currently holds
// priority within the step — added in S07 so the per-seat priority
// indicator is meaningful. PriorityHolder equals ActiveSeat at the
// start of every step that grants priority, or NoPriority during
// Untap and Cleanup.
type Turn struct {
	Seq            int // 1-indexed identity; increments every turn
	Round          int `json:"Number"` // 1-indexed table rotation; legacy snapshot key
	ActiveSeat     int // 0-indexed seat
	PriorityHolder int // 0-indexed seat or NoPriority during Untap/Cleanup
	Phase          Phase
	Step           Step
	Extra          bool // created by an effect (CR 500.7)
	ExtraRef       int  // queued extra-turn identity; zero on normal turns
	OrderSeat      int  // normal-rotation seat this turn follows
}

// TurnStep names ONE step of ONE turn — the cursor's Turn.Seq and
// Turn.Step frozen together, so a later read can ask "is the game
// still standing there?" rather than only "which step is this?".
//
// The zero value names no step at all, which is what makes it usable
// as an OPTIONAL anchor on a struct that mostly does not carry one:
// PendingChoice.OwedInStep, the step a pay-or-else prompt must be
// answered in (upkeep_pay_unless.go, CR 500.4).
type TurnStep struct {
	// Turn is the turn sequence this anchor is about, 1-indexed like
	// Turn.Seq. Zero means the anchor names no step. The field keeps its
	// legacy name because it is already persisted inside prompt state.
	Turn int
	// Step is the step within that turn.
	Step Step
}

// NamesAStep reports whether this anchor names one — false for the
// zero value, which is the "no anchor" case every optional user of
// the type has to be able to tell apart from a real step.
func (ts TurnStep) NamesAStep() bool { return ts.Turn > 0 }

// IsCurrent reports whether `cursor` is still standing in the step
// this anchor names. Always false for an anchor that names no step,
// so an unanchored caller gets "no" rather than an accidental match
// on turn zero.
func (ts TurnStep) IsCurrent(cursor Turn) bool {
	return ts.NamesAStep() && cursor.Seq == ts.Turn && cursor.Step == ts.Step
}

// stepGrantsPriority reports whether the given step grants priority
// to the active player on entry (CR 117). Untap and Cleanup are the
// only steps that do not.
func stepGrantsPriority(s Step) bool {
	return s != StepUntap && s != StepCleanup
}

// initialPriorityHolder returns the PriorityHolder value the cursor
// should hold immediately after entering step s with the given active
// seat — NoPriority for steps that don't grant priority, otherwise
// the active seat itself.
func initialPriorityHolder(s Step, activeSeat int) int {
	if !stepGrantsPriority(s) {
		return NoPriority
	}
	return activeSeat
}

// newStartingTurn returns the turn cursor at the start of a game:
// turn 1, the seat that won the opening roll, beginning phase, untap step. PriorityHolder is
// NoPriority because Untap doesn't grant priority. After the mulligan
// window closes, KeepHand fires runStepEntryHooksLocked which auto-
// untaps and advances the cursor into Upkeep.
func newStartingTurn(startingSeat int) Turn {
	return Turn{
		Seq:            1,
		Round:          1,
		ActiveSeat:     startingSeat,
		PriorityHolder: initialPriorityHolder(StepUntap, startingSeat),
		Phase:          PhaseOf(StepUntap),
		Step:           StepUntap,
		OrderSeat:      startingSeat,
	}
}

// indexOfStep returns the index of s within turnSequence, or -1 if s
// is not a recognised step.
func indexOfStep(s Step) int {
	for i, step := range turnSequence {
		if step == s {
			return i
		}
	}
	return -1
}

// IsNewTurn reports whether next identifies a different turn. Comparing
// identities rather than seats is what makes consecutive turns by the
// same player reset per-turn state correctly.
func (t Turn) IsNewTurn(next Turn) bool {
	return t.Seq != next.Seq
}

// advance returns the Turn cursor one step after t, wrapping to the
// next seat (and incrementing Seq, plus Round when the rotation returns to
// roundStartSeat) after the cleanup step. numSeats must be > 0; callers are
// responsible for passing a valid count and roundStartSeat.
//
// PriorityHolder is set via initialPriorityHolder so that landing on
// Untap or Cleanup yields NoPriority (S13 — those steps don't grant
// priority per CR 502.4 / 514.3).
func (t Turn) advance(numSeats, roundStartSeat int) Turn {
	idx := indexOfStep(t.Step)
	next := idx + 1
	if next < len(turnSequence) {
		nextStep := turnSequence[next]
		return Turn{
			Seq:            t.Seq,
			Round:          t.Round,
			ActiveSeat:     t.ActiveSeat,
			PriorityHolder: initialPriorityHolder(nextStep, t.ActiveSeat),
			Phase:          PhaseOf(nextStep),
			Step:           nextStep,
			Extra:          t.Extra,
			ExtraRef:       t.ExtraRef,
			OrderSeat:      t.OrderSeat,
		}
	}
	// Wrap: past cleanup, move to the next seat's untap. A table-facing
	// round is one full rotation from the player who started the game,
	// not from seat 0; join order must not leak back into turn display.
	nextSeat := (t.ActiveSeat + 1) % numSeats
	nextRound := t.Round
	if nextSeat == roundStartSeat {
		nextRound++
	}
	return Turn{
		Seq:            t.Seq + 1,
		Round:          nextRound,
		ActiveSeat:     nextSeat,
		PriorityHolder: initialPriorityHolder(StepUntap, nextSeat),
		Phase:          PhaseOf(StepUntap),
		Step:           StepUntap,
		OrderSeat:      nextSeat,
	}
}
