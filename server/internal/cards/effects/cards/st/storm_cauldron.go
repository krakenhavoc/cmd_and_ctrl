package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Storm Cauldron — Artifact {5} (Alliances, S82 OLD batch 820):
//
//	"Each player may play an additional land during each of their turns.
//	 Whenever a land is tapped for mana, return it to its owner's hand."
//
// The return is a triggered ability, so the mana is already in the pool when
// the land goes (the trigger resolves after it).
func init() {
	effects.Register(effects.Spec{
		OracleID:                         "5a51b168-02d1-4eb4-8ccc-614ab6f6cffa",
		Name:                             "Storm Cauldron",
		Completeness:                     effects.CompletenessFull,
		Purpose:                          game.Purpose{ExtraLandDrops: 1},
		AdditionalLandPlays:              1,
		AdditionalLandPlaysForEachPlayer: true,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventManaAbilityActivated},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B20LandTappedForMana(ev, g)
			},
			Key: "Storm Cauldron — return that land to its owner's hand",
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil || !effects.OnBattlefield(g, item.Trigger.Event.Source) {
					return nil
				}
				return effects.BounceToHand{Target: item.Trigger.Event.Source}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
