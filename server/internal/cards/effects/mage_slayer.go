package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mage Slayer — Artifact — Equipment {1}{R}{G}:
//
//	"Whenever equipped creature attacks, it deals damage equal to its
//	 power to the player or planeswalker it's attacking.
//	 Equip {3}"
//
// The trigger is Argentum Armor's (attachedCreatureAttacked), and it
// resolves in the declare attackers step; the creature still deals
// its combat damage later (the 2009-05-01 rulings).
//
// The EQUIPPED CREATURE deals the damage, not the Equipment, so its
// lifelink and deathtouch apply. It is named by object
// (Trigger.Object, Warstorm Surge's shape), and its power is read as
// the trigger resolves: live while it is on the battlefield, its
// last-known power once it has left (CR 608.2h). That is why the
// damage still happens if Mage Slayer has left the battlefield or
// moved to another creature by then (the ruling).
//
// The defender is the one that attacker was declared against, read
// off the attack event. A creature attacking a battle hits nothing,
// as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "db444262-c40f-4c80-9d14-265c32f77cf1",
		Name:         "Mage Slayer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attachedCreatureAttacked(ev, source)
			}, "Mage Slayer — equipped creature deals damage equal to its power to the player or planeswalker it's attacking",
				mageSlayerDamage),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}"),
		},
	})
}

// mageSlayerDamage is the trigger body.
func mageSlayerDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	creature, ok := ctx.TriggeringPermanent()
	if !ok {
		return nil
	}
	attacked := item.Trigger.Event.Target
	if g.PlayerByIDForEffect(attacked) == nil {
		if c, ok := g.LookupCardForEffect(attacked); !ok || !c.IsPlaneswalker() {
			return nil
		}
	}
	ref := ctx.Trigger().Object.Ref()
	return DealDamage{
		SourceObject: &ref,
		Target:       attacked,
		Amount:       creature.Power,
	}.Apply(ctx)
}
