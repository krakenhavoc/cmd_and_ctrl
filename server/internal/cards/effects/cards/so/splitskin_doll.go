package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Splitskin Doll — Artifact Creature — Toy {1}{W}, 2/1 (Duskmourn: House
// of Horror, #4466):
//
//	"When this creature enters, draw a card. Then discard a card unless
//	 you control another creature with power 2 or less."
//
// Seven of Nine's "draw a card. Then discard a card unless …": the board
// is read as the trigger resolves, and with another creature of power 2
// or less it is a plain draw, otherwise a loot (lootOne). Drawing a card
// changes nothing on the battlefield, so reading the board before the
// draw reads what "then" would. "Another" is object identity: the Doll
// itself never counts, even though it is a 2/1, and a Doll that has left
// counts nothing. Power is the creature's power now, pumps included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "934e2634-4b1c-4d92-b4d6-9241620f3770",
		Name:         "Splitskin Doll",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Splitskin Doll — draw a card, then discard a card unless you control another creature with power 2 or less",
				func(g *game.Game, item *game.StackItem) error {
					if splitskinDollAnotherSmallCreature(g, item.Controller, item.SourceCardID) {
						return effects.DrawCards{N: 1}.Apply(effects.NewContext(g, item))
					}
					return effects.LootOne(g, item, 1)
				}),
		},
	})
}

// splitskinDollAnotherSmallCreature is "you control another creature with
// power 2 or less": a creature `you` control, other than `self`, whose
// power is 2 or less.
func splitskinDollAnotherSmallCreature(g *game.Game, you, self uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != self && c.Controller == you && c.IsCreature() && c.CurrentPower() <= 2 {
			return true
		}
	}
	return false
}
