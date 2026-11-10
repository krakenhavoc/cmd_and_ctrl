package game

import "github.com/google/uuid"

// dealt_damage_mark.go — "has dealt damage" (Karakyk Guardian).
//
// markDealtDamageLocked records that a permanent has dealt damage of any
// kind (Card.DealtDamage), the twin of markDealtCombatDamageLocked, and,
// for a permanent with a static behind the HasntDealtDamage gate,
// schedules a layer pass, because the gate changes which of its statics
// are in play. A source that is not on the battlefield (it left before
// the damage settled) has no object to mark. Idempotent: only the first
// marking schedules the pass.
//
// Caller must hold g.mu.
func (g *Game) markDealtDamageLocked(source uuid.UUID) {
	c := g.findBattlefieldCardLocked(source)
	if c == nil || c.DealtDamage {
		return
	}
	c.DealtDamage = true
	if CatalogStaticAbilities == nil {
		return
	}
	for _, st := range CatalogStaticAbilities(catalogAbilityKeyOf(c)) {
		if st.ActiveWhen.Kind == DesignationHasntDealtDamage {
			g.layerVersion.Add(1)
			return
		}
	}
}
