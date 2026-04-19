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
// The twelve-step sequence at S02 is the complete MTG turn structure.
// Rules automation is not in scope here — advancing the step is a pure
// cursor move that the client drives via AdvanceStep.
type Step string

const (
	StepUntap            Step = "untap"
	StepUpkeep           Step = "upkeep"
	StepDraw             Step = "draw"
	StepPrecombatMain    Step = "precombat_main"
	StepBeginCombat      Step = "begin_combat"
	StepDeclareAttackers Step = "declare_attackers"
	StepDeclareBlockers  Step = "declare_blockers"
	StepCombatDamage     Step = "combat_damage"
	StepEndCombat        Step = "end_combat"
	StepPostcombatMain   Step = "postcombat_main"
	StepEnd              Step = "end"
	StepCleanup          Step = "cleanup"
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
	StepUntap:            PhaseBeginning,
	StepUpkeep:           PhaseBeginning,
	StepDraw:             PhaseBeginning,
	StepPrecombatMain:    PhasePrecombatMain,
	StepBeginCombat:      PhaseCombat,
	StepDeclareAttackers: PhaseCombat,
	StepDeclareBlockers:  PhaseCombat,
	StepCombatDamage:     PhaseCombat,
	StepEndCombat:        PhaseCombat,
	StepPostcombatMain:   PhasePostcombatMain,
	StepEnd:              PhaseEnding,
	StepCleanup:          PhaseEnding,
}

// PhaseOf returns the phase that contains a given step.
func PhaseOf(s Step) Phase {
	return stepPhase[s]
}

// Turn is the cursor into the game's turn/phase/step state machine.
// It says whose turn it is, which phase and step we're in, and the
// turn number since the game started (turn 1 = the first player's
// first turn). PriorityHolder tracks which seat currently holds
// priority within the step — added in S07 so the per-seat priority
// indicator is meaningful. PriorityHolder always equals ActiveSeat
// at the start of every step.
type Turn struct {
	Number         int // 1-indexed
	ActiveSeat     int // 0-indexed seat
	PriorityHolder int // 0-indexed seat; equals ActiveSeat at step boundaries
	Phase          Phase
	Step           Step
}

// newStartingTurn returns the turn cursor at the start of a game:
// turn 1, seat 0, beginning phase, untap step. Note that MTG's turn-1
// rules actually skip the draw step for the first player; S02 does not
// enforce this — manual play can account for it, and S13+ rules work
// will automate it.
func newStartingTurn() Turn {
	return Turn{
		Number:         1,
		ActiveSeat:     0,
		PriorityHolder: 0,
		Phase:          PhaseOf(StepUntap),
		Step:           StepUntap,
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

// advance returns the Turn cursor one step after t, wrapping to the
// next seat (and incrementing Number) after the cleanup step. numSeats
// must be > 0; callers are responsible for passing a valid count.
func (t Turn) advance(numSeats int) Turn {
	idx := indexOfStep(t.Step)
	next := idx + 1
	if next < len(turnSequence) {
		return Turn{
			Number:         t.Number,
			ActiveSeat:     t.ActiveSeat,
			PriorityHolder: t.ActiveSeat,
			Phase:          PhaseOf(turnSequence[next]),
			Step:           turnSequence[next],
		}
	}
	// Wrap: past cleanup, move to the next seat's untap. In Commander
	// with four players, after seat 3's cleanup we wrap to seat 0 and
	// the turn number goes up.
	nextSeat := (t.ActiveSeat + 1) % numSeats
	nextNumber := t.Number
	if nextSeat == 0 {
		nextNumber++
	}
	return Turn{
		Number:         nextNumber,
		ActiveSeat:     nextSeat,
		PriorityHolder: nextSeat,
		Phase:          PhaseOf(StepUntap),
		Step:           StepUntap,
	}
}
