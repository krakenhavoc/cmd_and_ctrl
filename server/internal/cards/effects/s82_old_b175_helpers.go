package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB175DamageToPlayerOrPlaneswalker deals `n` damage from the ability's source
// to `defender` when it names a player or a planeswalker on the battlefield, and
// does nothing for anything else (a battle is neither: "the player or
// planeswalker it's attacking").
func OldB175DamageToPlayerOrPlaneswalker(g *game.Game, item *game.StackItem, defender uuid.UUID, n int) error {
	if defender == uuid.Nil {
		return nil
	}
	if g.PlayerByIDForEffect(defender) == nil {
		c, ok := g.LookupCardForEffect(defender)
		if !ok || !c.IsPlaneswalker() || !g.Battlefield.Contains(defender) {
			return nil
		}
	}
	return DealDamage{Source: item.SourceCardID, Target: defender, Amount: n}.Apply(NewContext(g, item))
}

// OldB175SmallCreatureYouControlAttacks is "a creature you control with power
// `max` or less attacks": the attacker's power is read as it is declared.
func OldB175SmallCreatureYouControlAttacks(max int) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if !AttackDeclaredByYou(ev, source.Controller) {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && c.IsCreature() && c.CurrentPower() <= max
	}
}

// OldB175HitWhatItAttacks is "it deals `n` damage to the player or planeswalker
// it's attacking": the source's attack target as it resolves. A creature that
// has left combat is attacking nothing, and a battle takes no damage.
func OldB175HitWhatItAttacks(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		c, ok := g.LookupCardForEffect(item.SourceCardID)
		if !ok {
			return nil
		}
		return OldB175DamageToPlayerOrPlaneswalker(g, item, c.AttackingTarget, n)
	}
}

// OldB175HitTheAttackedPlayerOrPlaneswalker is the attack trigger's body for "this
// deals `n` damage to the player or planeswalker that creature is attacking":
// the attack event's target.
func OldB175HitTheAttackedPlayerOrPlaneswalker(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		return OldB175DamageToPlayerOrPlaneswalker(g, item, item.Trigger.Event.Target, n)
	}
}
