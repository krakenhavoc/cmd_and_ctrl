package game

// untap_counter_removal.go — "If this creature would untap during your
// untap step, remove a +1/+1 counter from it instead. If you do, untap
// it." (#6095, Bewitching Leechcraft's granted ability). A CR 614.1a
// replacement of one permanent's untap during its controller's untap step
// (CR 502.3): the untap happens only if the counter is removed, and
// otherwise the permanent stays tapped.
//
// It is not a ReplacementEffect: an untap is not an event the CR 614
// pipeline sees (stun counters, the other untap replacement, are read in
// the same primitive, untapPermanentLocked). It is plain data on the
// permanent's definition, read off its CatalogAbilityKey, so a grant
// bundle carries it like any other granted ability (the recipient has it
// while it has the grant, a later "loses all abilities" takes it, CR
// 613.6) and a permanent that prints it would declare it the same way.
//
// Several instances each replace the untap that the one before it left
// (CR 614.5 bars an effect from applying twice to one event, not a second
// effect from applying to the event the first produced): with two, the
// first removes a counter and untaps, the second replaces that untap with
// removing another, so the permanent untaps only if every instance removed
// a counter, and each instance that found one removed it.
//
// Only during the untap step of the permanent's controller ("your untap
// step"): an untap effect, a sandbox untap or another player's untap-step
// permission (Seedborn Muse) is not replaced. A stun counter's replacement
// is applied first: it removes the stun counter and the untap does not
// happen, which leaves the +1/+1 counters alone — the order the affected
// player always prefers (CR 616.1), since the other leaves the permanent
// tapped too and costs a counter more.

// UntapCounterRemoval is one "if this would untap during your untap step,
// remove a <Counter> counter from it instead. If you do, untap it."
// instance. Counter is the counter kind ("+1/+1"); Label names the clause
// for logs.
type UntapCounterRemoval struct {
	Counter string
	Label   string
}

// untapCounterRemovalsLocked is the instances `c` has right now, read off
// the layered view's catalog key (own abilities and every grant). Caller
// must hold g.mu.
func untapCounterRemovalsLocked(c *Card) []UntapCounterRemoval {
	d := catalogDef(catalogAbilityKeyOf(c))
	if d == nil {
		return nil
	}
	return d.UntapCounterRemovals
}

// replaceUntapWithCounterRemovalLocked applies `c`'s untap-step counter
// removals to an untap about to happen, and reports whether the untap
// still happens. It is a no-op (true) outside the untap step of `c`'s
// controller and for a permanent with none. Caller must hold g.mu (write).
func (g *Game) replaceUntapWithCounterRemovalLocked(c *Card) bool {
	if g.Turn.Step != StepUntap || c.Controller != g.activePlayerIDLocked() {
		return true
	}
	removals := untapCounterRemovalsLocked(c)
	if len(removals) == 0 {
		return true
	}
	id := c.InstanceID
	for _, r := range removals {
		cur := findBattlefieldCard(g, id)
		if cur == nil || r.Counter == "" || cur.Counters[r.Counter] <= 0 {
			return false
		}
		_ = g.applyCounterLocked(id, r.Counter, -1)
	}
	return true
}
