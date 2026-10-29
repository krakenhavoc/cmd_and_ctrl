package game

// FightForEffect makes two battlefield creature objects fight. Both powers
// and controllers are read before damage or its replacements; the fight
// exists even when no damage can be dealt. A missing, phased-out, changed
// object or noncreature prevents the entire fight.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) FightForEffect(a, b ObjectRef) error {
	g.RecomputeLayersIfStaleLocked()
	ca, cb := findBattlefieldCard(g, a.ID), findBattlefieldCard(g, b.ID)
	if ca == nil || cb == nil || ca.ObjectEpoch != a.Epoch || cb.ObjectEpoch != b.Epoch || !ca.IsCreature() || !cb.IsCreature() {
		return nil
	}
	pa, pb := ca.CurrentPower(), cb.CurrentPower()
	controllerA, controllerB := ca.Controller, cb.Controller
	// A creature can fight itself; that is one damage event for twice
	// its power, not two separate applications of damage replacements.
	if a == b {
		g.EmitEvent(Event{Kind: EventFight, Actor: controllerA, Source: a.ID, CardID: a.ID, Target: a.ID, BatchSeq: g.eventSeq + 1})
		return g.DealDamageFromObjectForEffect(a, a.ID, 2*pa)
	}
	seq := g.eventSeq + 1
	g.EmitEvent(Event{Kind: EventFight, Actor: controllerA, Source: a.ID, CardID: a.ID, Target: b.ID, BatchSeq: seq})
	g.EmitEvent(Event{Kind: EventFight, Actor: controllerB, Source: b.ID, CardID: b.ID, Target: a.ID, BatchSeq: seq})
	return g.DamageInstanceForEffect(func() error {
		if err := g.DealDamageFromObjectForEffect(a, b.ID, pa); err != nil {
			return err
		}
		return g.DealDamageFromObjectForEffect(b, a.ID, pb)
	})
}
