package ws

import (
	"slices"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// loop_replay.go — ADR 0193 §5: replay an accepted loop proposal
// through the engine, as one commit (CR 732.2c: "the game advances to
// the last proposed ending point, with all game choices contained in
// the shortcut proposal having been taken").
//
// Built like ApplyBundle: one hold of the room lock, one pre-clone, one
// undo entry (stamped with the proposer, so it costs one of their undos,
// owner decision 8), one snapshot and one replay line. Every recorded
// step goes through actions.Dispatch with its recorded actor, the same
// validation a click gets, so a replay can never do what the players
// could not.
//
// A replay stops at the first divergence (CR 732.2a: no conditional
// actions) — a step the engine refuses, a reference that names nothing,
// a prompt the recording did not see or did not get, a board that did
// not come back after an iteration, the game ending, or the step budget
// — and rolls back to the end of the last complete iteration, so the
// table is never left mid-loop.

// ErrNoLoopDue is ReplayLoop with no accepted proposal to replay.
var ErrNoLoopDue = errSentinel("ws: no accepted loop proposal to replay")

// ReplayLoop replays the accepted loop proposal now, as one commit. The
// room already does this at the end of the commit whose answer
// accepted it (autoAnswerThenCaptureLocked); ReplayLoop is for a
// caller that set a proposal up outside a commit.
func (r *Room) ReplayLoop() (protocol.GameView, uint64, error) {
	view, seq, err := r.replayLoop()
	if err == nil {
		r.notify()
	}
	return view, seq, err
}

func (r *Room) replayLoop() (protocol.GameView, uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.Game.LoopReplayDue() {
		return protocol.GameView{}, 0, ErrNoLoopDue
	}
	r.replayLoopLocked()
	return r.captureLocked(true)
}

// loopReplayBudget is the room's step budget for one replay:
// game.MaxLoopReplayDispatches unless a test set one.
func (r *Room) loopReplayBudget() int {
	if r.loopBudget > 0 {
		return r.loopBudget
	}
	return game.MaxLoopReplayDispatches
}

// replayLoopLocked runs the accepted proposal. Caller MUST hold r.mu.
func (r *Room) replayLoopLocked() {
	pre := r.Game.Clone()
	rec, err := r.Game.BeginLoopReplay()
	if err != nil {
		return
	}
	end := rec.Proposal.End
	budget := r.loopReplayBudget()
	dispatched := 0
	var out game.LoopReplayOutcome
	for it := 1; it <= end.Iteration; it++ {
		n := len(rec.Steps)
		if it == end.Iteration {
			n = end.Step
		}
		if n == 0 {
			break
		}
		mark := r.Game.Clone()
		births := r.Game.NewLoopBirths()
		why := ""
		for s := 0; s < n && why == ""; s++ {
			if dispatched >= budget {
				why = "the replay reached its limit of steps"
				break
			}
			dispatched++
			why = r.replayLoopStepLocked(rec.Steps[s], births)
		}
		if why == "" && r.Game.CurrentState() != game.StateActive {
			why = "the game ended"
		}
		if why == "" && n == len(rec.Steps) {
			if gap := r.Game.LoopIterationGap(rec); gap != "" {
				why = "the board did not come back: " + gap
			}
		}
		if why != "" {
			r.rollbackLocked(mark)
			out.Stopped = why
			break
		}
		if n == len(rec.Steps) {
			out.Completed++
		} else {
			out.Partial = true
		}
	}
	if out.Completed == 0 && !out.Partial {
		// Nothing stands: the replay commits nothing but its line.
		r.rollbackLocked(pre)
		r.Game.FinishLoopReplay(rec, out)
		return
	}
	r.Game.FinishLoopReplay(rec, out)
	r.pushUndoLocked(undoEntry{pre: pre, caller: rec.Recorder})
}

// replayLoopStepLocked dispatches one recorded step in the current
// iteration and answers what standing answers cover after it. It
// returns "" on success and the divergence otherwise. Caller MUST hold
// r.mu.
func (r *Room) replayLoopStepLocked(step game.LoopStep, births *game.LoopBirths) string {
	births.Scan(r.Game)
	if got := r.Game.LoopPendingShape(); !slices.Equal(got, step.Pending) {
		if len(got) > len(step.Pending) {
			return "a prompt came up that the recording has no answer for, before “" + step.Label + "”"
		}
		return "a prompt the recording answered was not asked, at “" + step.Label + "”"
	}
	a, err := actions.ResolveLoopStep(step, births)
	if err != nil {
		return "“" + step.Label + "” names something this iteration did not create"
	}
	if err := actions.Dispatch(r.Game, a); err != nil {
		_, msg := classifyActionError(err)
		return "“" + step.Label + "” was refused: " + strings.TrimSuffix(msg, ".")
	}
	r.settleAutoAnswersLocked()
	return ""
}

// settleAutoAnswersLocked answers every prompt a seat has a standing
// answer for (ADR 0127), as autoAnswerThenCaptureLocked does, but
// inside a replay: no frames between them and no undo entries, since
// the replay is one commit. Caller MUST hold r.mu.
func (r *Room) settleAutoAnswersLocked() {
	r.Game.SettlePassTurn()
	for i := 0; i < maxAutoAnswersPerCommit; i++ {
		choiceID, _, ok := r.Game.NextAutoAnswer()
		if !ok {
			return
		}
		if _, err := r.Game.AutoAnswer(choiceID); err != nil {
			r.Game.MarkAskedByHand(choiceID, game.AskedByHandUndone)
			continue
		}
		r.Game.SettlePassTurn()
	}
}
