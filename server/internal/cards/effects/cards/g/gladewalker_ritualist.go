package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gladewalker Ritualist — Creature — Shapeshifter {2}{G}, 3/3:
//
//	"Changeling (This card is every creature type.)
//	 Whenever another creature you control named Gladewalker Ritualist
//	 enters, draw a card."
//
// The name is read off the entering permanent as it is now (a copy of the
// Ritualist is named Gladewalker Ritualist too); "another" leaves out the
// Ritualist the ability is on. Two Ritualists entering together each see the
// other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "37420ccd-c6c6-4f15-a696-d8986934afb5",
		Name:            "Gladewalker Ritualist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"changeling"},
		Purpose:         game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature() && c.Name == "Gladewalker Ritualist"
			}, "Gladewalker Ritualist — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
