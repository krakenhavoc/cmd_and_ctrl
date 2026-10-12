package game

// copyTargetsCanFinishLocked lets the sequential copy picker change both
// members of a related pair (CR 115.7e: only the final set is judged).
// Ordinary copies keep their existing gate. With a relation, only future
// dependent slots need exploration; every anchor is already a required
// single pick and a dependent clause has at most one pick.
func (g *Game) copyTargetsCanFinishLocked(src TargetSource, steps []AnnouncedClause, original, next []TargetRef, after int) bool {
	trial := append([]TargetRef(nil), next...)
	var finish func(int) bool
	finish = func(index int) bool {
		for index < len(steps) && steps[index].Clause.Relation == nil {
			index++
		}
		if index == len(steps) {
			return g.retargetCheckLocked(src, steps, original, trial, RetargetChooseNew) == nil
		}
		step := steps[index]
		position := -1
		for i, ref := range trial {
			if ref.Mode == step.Mode && ref.Slot == step.Slot {
				position = i
				break
			}
		}
		if position < 0 {
			return finish(index + 1)
		}
		previous := trial[position]
		if finish(index + 1) {
			return true
		}
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
	return finish(after)
}
