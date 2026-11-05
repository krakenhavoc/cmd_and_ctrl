package game

import "github.com/google/uuid"

// HealAllDamageForEffect is "all damage already dealt to it is healed"
// (CR 701.69a): every point of damage marked on the permanent is removed
// (Dr. McCoy, Surly Surgeon, #4973). Unlike HealOtherDamageForEffect it
// is not tied to a damage event, so nothing marked is kept. The
// deathtouch mark is a fact about damage dealt rather than marked damage
// (heal_damage.go) and is left to consumeDeathtouchMarksLocked. Returns
// how much was healed; zero when the card is not on the battlefield.
//
// Caller must hold g.mu (write).
func (g *Game) HealAllDamageForEffect(cardID uuid.UUID) int {
	c := findBattlefieldCard(g, cardID)
	if c == nil || c.DamageMarked <= 0 {
		return 0
	}
	n := c.DamageMarked
	c.DamageMarked = 0
	c.damageInstance, c.damageInstanceMarked = 0, 0
	return n
}
