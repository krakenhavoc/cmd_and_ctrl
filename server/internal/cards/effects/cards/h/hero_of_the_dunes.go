package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hero of the Dunes — Creature — Human Soldier {3}{W}{B}, 3/2 (The Brothers'
// War, #5482):
//
//	"When this creature enters, return target artifact or creature card with
//	 mana value 3 or less from your graveyard to the battlefield.
//	 Creatures you control with mana value 3 or less get +1/+0."
//
// The static reads each creature's mana value through the layered view as
// the layers are computed, so a token (mana value 0) is covered and a
// creature that copies something bigger is not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eb6f936e-8dba-49da-84fc-de240b11b8e4",
		Name:         "Hero of the Dunes",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.Labelled(effects.B16Anthem(func(target *game.Card, g *game.Game, source *game.Card) bool {
				if !target.IsCreature() || target.Controller != source.Controller {
					return false
				}
				mv, ok := g.ManaValueForEffect(*target)
				return ok && mv <= 3
			}, 1, 0), "Creatures you control with mana value 3 or less get +1/+0."),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Hero of the Dunes — return target artifact or creature card with mana value 3 or less from your graveyard to the battlefield",
					effects.S82Bro10ReturnTargetToBattlefield(false)),
				effects.TargetCardInGraveyard("target artifact or creature card with mana value 3 or less from your graveyard",
					effects.YouOwn(), effects.Or(effects.Artifact(), effects.Creature()), effects.ManaValueLE(3)),
			),
		},
	})
}
