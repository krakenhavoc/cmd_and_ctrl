package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shared Animosity — Enchantment {2}{R}:
//
//	"Whenever a creature you control attacks, it gets +1/+0 until end
//	 of turn for each other attacking creature that shares a creature
//	 type with it."
//
// One trigger per attacking creature (the text is "a creature", not
// "one or more"). It reads the attacker off the
// triggering EventAttack (item.Trigger.Event.CardID) and counts at
// RESOLUTION, so an attacker that has since left combat no longer
// counts (CR 608.2h). "Other attacking creature" is any attacker
// anyone controls, not only yours: the printed text names no
// controller, so an attacking creature of another player that shares
// a type with it counts too. game.SharesCreatureType is the type test,
// so a changeling shares with every creature that has a creature type.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a27445db-33f2-4571-98b5-83206b797484",
		Name:         "Shared Animosity",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Shared Animosity — +1/+0 for each other attacker that shares a creature type",
				sharedAnimosityBoost),
		},
	})
}

// sharedAnimosityBoost pumps the triggering attacker by the number of
// other attacking creatures that share a creature type with it.
func sharedAnimosityBoost(g *game.Game, item *game.StackItem) error {
	id := item.Trigger.Event.CardID
	attacker, ok := g.LookupCardForEffect(id)
	if !ok || attacker.AttackingTarget == uuid.Nil {
		return nil
	}
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == id || c.AttackingTarget == uuid.Nil {
			continue
		}
		if game.SharesCreatureType(&attacker, &c) {
			n++
		}
	}
	return BoostUntilEOT{
		Target: id,
		Power:  n,
		Label:  "Shared Animosity — +1/+0 for each other attacking creature that shares a creature type",
	}.Apply(NewContext(g, item))
}
