package game

import "github.com/google/uuid"

// copyTargetsCanFinishLocked lets the sequential copy picker change both
// members of a related pair (CR 115.7e: only the final set is judged).
// Ordinary copies keep their existing gate. With a relation, only future
// dependent slots need exploration: every anchor is already a required
// single pick. The steps before `after` are answered; the picks of the
// steps from `after` on may still change.
func (g *Game) copyTargetsCanFinishLocked(src TargetSource, steps []AnnouncedClause, original, next []TargetRef, after int) bool {
	if after < 0 {
		after = 0
	}
	open := make(map[int]bool)
	for i, ref := range next {
		for _, step := range steps[min(after, len(steps)):] {
			if ref.Mode == step.Mode && ref.Slot == step.Slot {
				open[i] = true
			}
		}
	}
	return g.relatedTargetsCanFinishLocked(src, steps, original, next, after, func(i int) bool { return open[i] })
}

// targetsAfterCanFinishLocked is copyTargetsCanFinishLocked for a walk
// that asks one pick at a time (a related retarget): the picks at
// positions up to and including `pos` are answered, and every later one,
// the rest of pos's own clause included (#4862), may still change.
func (g *Game) targetsAfterCanFinishLocked(src TargetSource, steps []AnnouncedClause, original, next []TargetRef, pos int) bool {
	from := len(steps)
	if pos >= 0 && pos < len(next) {
		for j, step := range steps {
			if step.Mode == next[pos].Mode && step.Slot == next[pos].Slot {
				from = j
				break
			}
		}
	}
	return g.relatedTargetsCanFinishLocked(src, steps, original, next, from, func(i int) bool { return i > pos })
}

// relatedTargetsCanFinishLocked reports whether the dependent clauses
// from steps[from] on can be completed, changing only the picks whose
// positions `open` admits, into a set the CR 115.7d gate accepts.
//
// A dependent clause with one open pick tries each candidate in turn. One
// with several open picks (a ManyPicks relation, #4862) is refilled
// rather than searched: a pick its anchor still admits stays, and each
// other open pick takes the next candidate the clause does not already
// hold. That is exact because such a relation judges every pick against
// the anchor alone, and Register keeps set rules off the clause, so any
// distinct candidates fit as well as any others; a pick with no candidate
// left keeps its object and the final check decides.
func (g *Game) relatedTargetsCanFinishLocked(src TargetSource, steps []AnnouncedClause, original, next []TargetRef, from int, open func(int) bool) bool {
	trial := append([]TargetRef(nil), next...)
	var finish func(int) bool
	finish = func(index int) bool {
		for index < len(steps) && steps[index].Clause.Relation == nil {
			index++
		}
		if index >= len(steps) {
			return g.retargetCheckLocked(src, steps, original, trial, RetargetChooseNew) == nil
		}
		step := steps[index]
		var positions []int
		for i, ref := range trial {
			if open(i) && ref.Mode == step.Mode && ref.Slot == step.Slot {
				positions = append(positions, i)
			}
		}
		if finish(index + 1) {
			return true
		}
		switch len(positions) {
		case 0:
			return false
		case 1:
			position := positions[0]
			previous := trial[position]
			lt := g.LegalTargetsForClauseForEffect(src, step, trial)
			for _, id := range lt.Cards {
				trial[position] = TargetRef{Kind: TargetCard, ID: id, Mode: step.Mode, Slot: step.Slot}
				if finish(index + 1) {
					return true
				}
			}
			trial[position] = previous
			return false
		}
		return g.refillRelatedPicksLocked(src, step, trial, positions, func() bool { return finish(index + 1) })
	}
	return finish(from)
}

// refillRelatedPicksLocked is relatedTargetsCanFinishLocked's step for a
// dependent clause with several open picks. It leaves trial as it found
// it unless `rest` succeeds.
func (g *Game) refillRelatedPicksLocked(src TargetSource, step AnnouncedClause, trial []TargetRef, positions []int, rest func() bool) bool {
	isOpen := make(map[int]bool, len(positions))
	for _, p := range positions {
		isOpen[p] = true
	}
	// The clause's legal set beside every pick but its open ones: the
	// anchor and the clause's answered picks still count.
	others := make([]TargetRef, 0, len(trial))
	held := make(map[uuid.UUID]bool)
	for i, ref := range trial {
		if isOpen[i] {
			continue
		}
		others = append(others, ref)
		if ref.Mode == step.Mode && ref.Slot == step.Slot {
			held[ref.ID] = true
		}
	}
	lt := g.LegalTargetsForClauseForEffect(src, step, others)
	candidates := make([]TargetRef, 0, len(lt.Players)+len(lt.Cards))
	legal := make(map[uuid.UUID]bool, cap(candidates))
	for _, id := range lt.Players {
		candidates = append(candidates, TargetRef{Kind: TargetPlayer, ID: id, Mode: step.Mode, Slot: step.Slot})
		legal[id] = true
	}
	for _, id := range lt.Cards {
		candidates = append(candidates, TargetRef{Kind: TargetCard, ID: id, Mode: step.Mode, Slot: step.Slot})
		legal[id] = true
	}
	saved := make([]TargetRef, len(positions))
	var refill []int
	for k, p := range positions {
		saved[k] = trial[p]
		if id := trial[p].ID; legal[id] && (step.Clause.AllowSame || !held[id]) {
			held[id] = true
			continue
		}
		refill = append(refill, p)
	}
	if len(refill) == 0 {
		return false
	}
	next := 0
	for _, p := range refill {
		for next < len(candidates) && !step.Clause.AllowSame && held[candidates[next].ID] {
			next++
		}
		if next == len(candidates) {
			break
		}
		trial[p] = candidates[next]
		held[candidates[next].ID] = true
		next++
	}
	if rest() {
		return true
	}
	for k, p := range positions {
		trial[p] = saved[k]
	}
	return false
}
