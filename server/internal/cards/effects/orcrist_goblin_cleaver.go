package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Orcrist, Goblin-cleaver — Legendary Artifact — Equipment {3} (The Hobbit):
//
//	"Equipped creature gets +2/+2 and has trample.
//	 Whenever equipped creature deals combat damage to a player, choose a
//	 creature type. Create a Treasure token for each creature you control of
//	 that type.
//	 Equip {3}"
//
// #4454, S76. Bloodforged Battle-Axe's equipped-creature combat damage
// trigger with Distant Melody's choose-a-type tail. The tokens are counted
// when the type has been chosen, creatures only (a Treasure that is also
// that type does not count), and a changeling is of every type.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cc65821a-1893-4087-b461-35cf9fd26c71",
		Name:         "Orcrist, Goblin-cleaver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("trample"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Key: "Orcrist — choose a creature type, then a Treasure for each creature you control of that type",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ChooseCreatureTypeTail(g, item.Controller, item.SourceCardID, "Orcrist, Goblin-cleaver — choose a creature type",
					NewContext(g, item).ItemTail(orcristTypeChosen))
				return nil
			},
		}},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}

// orcristTypeChosen is "create a Treasure token for each creature you control
// of that type". An on-disk key.
var orcristTypeChosen = creatureTypeChosen("then/orcrist-type-chosen", func(ctx *Context, t string, _ game.TailOutcome) error {
	n := len(creaturesOfTypeYouControl(ctx.Game, ctx.Controller(), t))
	if n == 0 {
		return nil
	}
	return CreateToken{Controller: ctx.Controller(), Template: TreasureToken(), N: n}.Apply(ctx)
})
