package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Generous Patron — Creature — Elf Advisor {2}{G}, 1/4:
//
//	"When this creature enters, support 2. (Put a +1/+1 counter on each
//	 of up to two other target creatures.)
//	 Whenever you put one or more counters on a creature you don't
//	 control, draw a card."
//
// Support is a permanent's here, so the targets are other creatures
// (CR 701.41a). The draw is once per creature per placement
// (WheneverYouPutOneOrMoreCountersOnACreature): supporting two of your
// opponents' creatures draws two cards, and putting two kinds of
// counter on one creature at once draws one. Proliferate counts,
// as does any counter of any kind.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb084141-cb80-479e-b5d4-114c1f31eb6c",
		Name:         "Generous Patron",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersSupport("Generous Patron", 2, true),
			effects.WheneverYouPutOneOrMoreCountersOnACreature(effects.CreatureYouDontControl,
				"Generous Patron — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
