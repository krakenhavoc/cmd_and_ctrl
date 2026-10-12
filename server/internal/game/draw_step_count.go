package game

import "github.com/google/uuid"

// FirstDrawInOwnDrawStepForEffect asks about an actual draw, not a draw
// instruction: replacing the first draw with milling leaves the exception
// available for the next draw. Caller holds g.mu.
func (g *Game) FirstDrawInOwnDrawStepForEffect(player uuid.UUID) bool {
	return g.Turn.Step == StepDraw && g.Turn.ActiveSeat >= 0 && g.Turn.ActiveSeat < len(g.Seats) &&
		g.Seats[g.Turn.ActiveSeat] != nil && g.Seats[g.Turn.ActiveSeat].ID == player && g.DrawnThisStep[player] == 0
}

// FirstCompletedDrawInOwnDrawStepForEffect is the corresponding test for
// a draw trigger. The actual draw has already incremented the tally when
// its event is emitted.
func (g *Game) FirstCompletedDrawInOwnDrawStepForEffect(player uuid.UUID) bool {
	return g.Turn.Step == StepDraw && g.Turn.ActiveSeat >= 0 && g.Turn.ActiveSeat < len(g.Seats) &&
		g.Seats[g.Turn.ActiveSeat] != nil && g.Seats[g.Turn.ActiveSeat].ID == player && g.DrawnThisStep[player] == 1
}

// A pre-v8 restore has no tally. The step boundary and completed draw
// events reconstruct it; DrawnThisTurn is a conservative fallback for an
// older log that no longer contains the current boundary.
func (g *Game) restoreDrawStepCountLocked() {
	if g.DrawnThisStep != nil || g.Turn.Step != StepDraw {
		return
	}
	g.DrawnThisStep = make(map[uuid.UUID]int)
	boundary := -1
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == EventStepBegan {
			boundary = i
			break
		}
	}
	if boundary >= 0 {
		for _, ev := range g.Events[boundary+1:] {
			if ev.Kind == EventDrawCard {
				g.DrawnThisStep[ev.Actor]++
			}
		}
	} else {
		for player, ids := range g.DrawnThisTurn {
			g.DrawnThisStep[player] = len(ids)
		}
	}
}
