package effects

// fightTargetClauses is a fight between the still-legal picks of the
// first two target clauses, shared by targeted fight spells (#4380).
func fightTargetClauses(ctx *Context) error {
	mine, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	theirs, ok := ctx.ClauseTarget(1)
	if !ok {
		return nil
	}
	return b10Fight(ctx, mine.ID, theirs.ID)
}
