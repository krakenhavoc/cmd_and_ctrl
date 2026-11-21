package game

// stampCombatSourceNameLocked records, on a combat EventDealDamage, the
// layered name of the object that dealt the damage (Event.SourceName). The
// dealing object may be a creature that has since left the game, so the name
// is captured as the damage is dealt. A name already on the event is kept.
// Caller must hold g.mu.
func (g *Game) stampCombatSourceNameLocked(ev *Event) {
	if ev.Kind != EventDealDamage || !ev.Combat || ev.SourceName != "" {
		return
	}
	z, pos := g.locateCardLocked(ev.Source)
	if z == nil || pos < 0 || pos >= len(z.Cards) {
		return
	}
	ev.SourceName = EffectiveOf(&z.Cards[pos]).Name
}
