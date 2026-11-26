package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avarice Amulet — Artifact — Equipment {4} (Magic 2015, #5556):
//
//	"Equipped creature gets +2/+0 and has vigilance and "At the beginning
//	 of your upkeep, draw a card."
//	 Whenever equipped creature dies, target opponent gains control of this
//	 Equipment.
//	 Equip {2}"
//
// The pump and vigilance are Accorder's Shield's statics; the upkeep draw is
// a granted triggered ability (ADR 0093), so its controller is the creature's
// controller. The dies trigger is Skullclamp's (the attachment is still read
// when the creature leaves), with a target opponent chosen for it; the
// Equipment changes hands for good (CR 611.2a, a layer-2 effect with no
// duration), so that opponent also owns the next Equip.
//
// No simplification.
const avariceAmuletGrant = "avarice-amulet/upkeep-draw"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c71074b-b8bf-47f4-87e4-67c275abdbc5",
		Name:         "Avarice Amulet",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: avariceAmuletGrant,
			Triggered: []game.TriggeredAbility{
				effects.AtYourUpkeep("Avarice Amulet — draw a card", effects.Do(effects.DrawCards{N: 1})),
			},
			Text: "At the beginning of your upkeep, draw a card.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 0),
			effects.GrantToAttached("vigilance"),
			effects.GrantAbilitiesToAttached(avariceAmuletGrant),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.EquippedCreatureDied(ev, source)
				}, "Avarice Amulet — target opponent gains control of this Equipment",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						targets := ctx.LegalTargets()
						if len(targets) == 0 || !effects.SourceIsStillThisPermanent(g, item) {
							return nil
						}
						return effects.GainControl{
							Target:     item.SourceCardID,
							Controller: targets[0].ID,
							Duration:   game.IndefiniteDuration(),
							Label:      "Avarice Amulet — target opponent gains control of it",
						}.Apply(ctx)
					}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
