package game

// trigger_mode_cast_x.go — #5898: a modal enters trigger whose chosen
// mode's target clause is sized and bounded by the X its source was
// cast with. Orcus, Prince of Undeath: "When Orcus enters, choose one
// — … • Return up to X target creature cards with total mana value X
// or less from your graveyard to the battlefield."
//
// CR 107.3m: "If an object's enters-the-battlefield triggered ability
// or replacement effect refers to X, and the spell that became that
// object as it resolved had a value of X chosen for any of its costs,
// the value of X for that ability is the same as the value of X for
// that spell, although the value of X for that permanent is 0."
//
// The non-modal form already exists (xCountedTriggerTargets, Lost in
// the Maze): a CountFromX clause on a XFromCast row. A modal trigger's
// clauses are its chosen modes', listed by AnnouncedClauses, so the
// same rewrite runs over those steps here, at the two places the
// steps are listed from the catalog: the CR 603.3d target walk's start
// (buildOrPickTriggerLocked) and the walk's rebuild on restore
// (rebuildTargetWalkLocked). Both read the X off the frame's source,
// the value copy of the permanent as it entered (Card.CastX), which a
// restore writes back with the frame — so the X survives a save.
//
// Then every X bound on those steps is written with the same X
// (bindStepsX): "with mana value X or less" and "with total mana value
// X or less" (TargetTotal.MaxIsX). Once the trigger is built, its item
// carries the X as XValue (XFromCast), and itemAnnouncedClauses binds
// the bounds from that for the CR 608.2b re-check, as for a spell.
//
// Data only: the row's XFromCast flag and the clause's flags, no func
// reachable from Game.

// bindTriggerModeStepsCastX rewrites a XFromCast modal trigger's
// chosen-mode steps for the X its source was cast with. A CountFromX
// step takes exactly X targets, or up to X for an UpToX one; at X = 0
// such a step takes none and is dropped (a Max of 0 would read as
// unbounded). Any other row, and a non-modal one (whose clause
// xCountedTriggerTargets has already sized), is returned as it was.
// The steps hold clause copies, so the catalog is never touched.
func bindTriggerModeStepsCastX(t TriggeredAbility, source *Card, steps []AnnouncedClause) []AnnouncedClause {
	if !t.XFromCast || t.Modes == nil || t.Targets != nil || len(steps) == 0 {
		return steps
	}
	x := 0
	if source != nil {
		x = source.CastX()
	}
	out := steps[:0:0]
	for _, s := range steps {
		if s.Clause.CountFromX {
			if x <= 0 {
				continue
			}
			s.Clause.Min, s.Clause.Max = x, x
			if s.Clause.UpToX {
				s.Clause.Min = 0
			}
			s.Clause.CountFromX = false
		}
		out = append(out, s)
	}
	bindStepsX(out, x)
	return out
}
