package game

// trigger_event_x.go — a trigger's "up to X targets" where X is the
// triggering event's amount (#6085).
//
//	Rampaging War Mammoth  Cycling {X}{2}{R}. When you cycle this card,
//	                       destroy up to X target artifacts.
//
// A triggered ability announces no X of its own. The X here is the one
// announced for the cycling cost (CR 107.3a), which EventCycle carries as
// its Amount (ADR 0062 Decision 7), and CR 603.3d chooses the trigger's
// targets as it is put on the stack, after that event, so the count is
// known then. A TriggeredAbility marked XFromEventAmount reads it there:
// triggerTargetsLocked rewrites a CountFromX clause into one of exactly X
// targets (up to X for an UpToX clause), on a copy so the catalog's clause
// is untouched — xCountedTriggerTargets' rewrite with the event's amount
// in place of the source's cast X. triggerTargetsLocked is the one place
// every reader lists the clause from (the harvest, the CR 603.3d walk and
// the walk's rebuild on restore, which keeps the event in its trigger
// context), so they all see the same count.
//
// X = 0 is no targets at all, never an unbounded clause: a Max of 0 reads
// as unbounded everywhere else in a clause (#259), so the clause is
// dropped and the ability has no target to choose and does nothing.
//
// Data only: the row's flag and the clause's, no func reachable from Game.

// eventCountedTriggerTargets is the CountFromX clause of an
// XFromEventAmount trigger counted by the triggering event's Amount, or
// nil when that is zero or less (the ability then has no targets).
func eventCountedTriggerTargets(spec *TargetSpec, ev Event) *TargetSpec {
	x := ev.Amount
	if x <= 0 {
		return nil
	}
	cp := *spec
	cp.Min, cp.Max, cp.CountFromX = x, x, false
	if spec.UpToX {
		cp.Min = 0
	}
	return &cp
}
