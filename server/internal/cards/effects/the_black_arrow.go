package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Black Arrow — Legendary Artifact — Equipment {3} (The Hobbit):
//
//	"Flash
//	 When The Black Arrow enters, it deals 1 damage to any target. If a
//	 Dragon is dealt damage this way, destroy it.
//	 Equipped creature gets +1/+1 and has reach.
//	 Equip {1}"
//
// #4454, S76. Ghitu Slinger's enters-damage ping, with Ballista Wielder's
// rider shape: a permanent hit is dealt its damage through a keyed tail
// (the Dragon check is made on what was actually dealt, so prevented damage
// does not destroy), and a player is dealt it plain. The +1/+1 and reach
// are PumpAttached and GrantToAttached.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cb2803ee-825b-420e-b404-c7c44f15ec15",
		Name:            "The Black Arrow",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("reach"),
		},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				Targeting(WhenThisEnters("The Black Arrow — 1 damage to any target; destroy it if it is a Dragon", hobB5BlackArrowShot), TargetAny()),
				ForTargets(DamageToTarget(0, 1))),
		},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}

func hobB5BlackArrowShot(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	t := targets[0]
	if t.Kind != game.TargetCard {
		return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 1}.Apply(ctx)
	}
	return g.DealDamageToCreatureKeyedForEffect(item.SourceCardID, t.ID, 1, ctx.ItemTail(hobB5BlackArrowDragon))
}

// hobB5BlackArrowDragon is "if a Dragon is dealt damage this way, destroy
// it", told the permanent and the damage it was dealt. "Dragon" is a word of
// this card's text (CR 612.2). An on-disk key.
var hobB5BlackArrowDragon = CardTail("then/the-black-arrow-destroy-dragon", func(ctx *Context, o game.TailOutcome) error {
	if o.Amount <= 0 || len(o.IDs) == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok || !c.HasSubtype(ctx.TextSubtype("Dragon")) {
		return nil
	}
	return nothingIfGone(ctx.Game.DestroyPermanentForEffect(o.IDs[0]))
})
