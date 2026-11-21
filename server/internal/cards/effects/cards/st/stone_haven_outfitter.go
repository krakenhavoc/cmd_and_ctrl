package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stone Haven Outfitter — {1}{W} Creature — Kor Artificer Ally 2/2
// (Foundations Jumpstart):
//
//	"Equipped creatures you control get +1/+1.
//	 Whenever an equipped creature you control dies, draw a card."
//
// The anthem reads "equipped" on every recompute (isEquippedCreature). The
// trigger asks whether the creature that died was equipped as it last
// existed (CR 603.10a), as Rhuk's does.
//
// No simplification.
func init() {
	equipped := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && effects.IsEquippedCreature(g, *target)
	}
	effects.Register(effects.Spec{
		OracleID:     "8e646a6b-77ba-4f7d-8ee9-b0b084453f77",
		Name:         "Stone Haven Outfitter",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.B16Anthem(equipped, 1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller && effects.RhukWasEquipped(g, ev.CardID)
			}, "Stone Haven Outfitter — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
