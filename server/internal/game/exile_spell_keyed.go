package game

import "github.com/google/uuid"

// exile_spell_keyed.go — #4253: "exile target spell" with the rest of the
// effect as data, so a spell's exit paused on a prompt (a CR 616 ordering of
// two replacements of the move, a "may" replacement's question) is a
// restore point.
//
// The closure form (ExileSpellThenForEffect) handed its caller's
// continuation to the exit route as a closure (zoneRoute.then), which a
// restore point cannot carry; #4253's cleanup deleted it once no caller was
// left. The keyed form rides the route as a one-leg route batch
// (zoneRoute.batchLegs, route_batch_keyed.go): the batch's leg is the spell,
// nothing is left to route after it, and the batch's continuation is the
// caller's registered tail, told the spell when it reached exile (CR 400.7)
// and nothing when it did not. The route is the one every spell exit builds
// (spellExitRouteLocked: the stack record retired, flashback's CR 702.34a
// exile applied).

// ExileSpellKeyedForEffect is ExileSpellForEffect with the rest of the
// effect as data: `then` is a registered tail (RegisterTail), told the spell as
// TailOutcome.IDs when it is in exile once the move settles, and no IDs when
// it is not (it went to the command zone instead, CR 903.9, or had already
// left the stack: countered or resolved in response, which is not an
// error). A zero tail is "no continuation", ExileSpellForEffect.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) ExileSpellKeyedForEffect(stackID uuid.UUID, then KeyedTail) error {
	if item, ok := g.StackMeta[stackID]; !ok || item == nil || item.Kind != StackItemSpell ||
		g.Stack == nil || !g.Stack.Contains(stackID) {
		return g.runKeyedTailLocked(then, nil, 0)
	}
	if then.IsZero() {
		return g.ExileSpellForEffect(stackID)
	}
	r, err := g.spellExitRouteLocked(stackID, &ZoneRef{Kind: ZoneExile}, ZoneExile, false, TuckOptions{}, nil)
	if err != nil {
		return err
	}
	r.batchLegs = &routeBatchRun{Leg: stackID, Then: snapshotKeyedTail(then)}
	_, err = g.routeCardToZoneLocked(r)
	return err
}
