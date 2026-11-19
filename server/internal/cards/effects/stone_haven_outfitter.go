package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
		return target.Controller == source.Controller && isEquippedCreature(g, *target)
	}
	Register(Spec{
		OracleID:     "8e646a6b-77ba-4f7d-8ee9-b0b084453f77",
		Name:         "Stone Haven Outfitter",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{b16Anthem(equipped, 1, 1)},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && leftUnderControlOf(ev, dead) == source.Controller && rhukWasEquipped(g, ev.CardID)
			}, "Stone Haven Outfitter — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
