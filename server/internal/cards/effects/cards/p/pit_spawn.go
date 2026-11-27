package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pit Spawn — Creature — Demon {4}{B}{B}{B}, 6/4 (Exodus, S82 OLD b738, #5626):
//
//	"First strike
//	 At the beginning of your upkeep, sacrifice this creature unless you pay
//	 {B}{B}.
//	 Whenever this creature deals damage to a creature, exile that creature."
//
// First strike is a printed keyword; the upkeep payment is the shared
// "sacrifice unless you pay" (Molting Harpy). The exile trigger is Spiritmonger's
// watch (any damage, combat or not) naming the damaged creature as the object
// it was (CR 400.7): a creature that has already left, or left and returned, is
// not exiled.
//
// No simplification.
func init() {
	const label = "Pit Spawn — exile that creature"
	effects.Register(effects.Spec{
		OracleID:        "f536acf3-ff8e-48da-8509-00e7ce4a1678",
		Name:            "Pit Spawn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.UpkeepSacrificeUnlessYouPay("Pit Spawn", "{B}{B}"),
			{
				Watches:   []game.EventKind{game.EventDealDamage},
				AppliesTo: effects.ThisDealtDamageToACreature,
				Key:       label,
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, label)
					item.Params.Object = effects.DamagedCreatureRef(g, ev.Target)
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					info, ok := g.PermanentForEffect(item.Params.Object)
					if !ok || info.Left {
						return nil
					}
					return effects.ExileTarget{Target: item.Params.Object.ID}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
