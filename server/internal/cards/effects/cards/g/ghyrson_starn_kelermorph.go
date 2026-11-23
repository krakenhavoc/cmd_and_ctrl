package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghyrson Starn, Kelermorph — Legendary Creature — Tyranid Human {1}{U}{R},
// 3/2 (Warhammer 40,000 Commander, S82 GAP batch 194, #5948):
//
//	"Ward {2}
//	 Three Autostubs — Whenever another source you control deals exactly 1
//	 damage to a permanent or player, Ghyrson Starn deals 2 damage to that
//	 permanent or player."
//
// One trigger per damage event of exactly 1 from another source of yours (a
// creature, a spell on the stack, a permanent's ability), read as the damage
// was dealt. Ghyrson's own damage never triggers it, so the 2 damage cannot
// loop. The 2 damage goes to the recipient the event named, which is still read
// as it is now: a permanent that has left takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b110b8e6-859b-4370-aceb-b21c6cf61532",
		Name:         "Ghyrson Starn, Kelermorph",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Ghyrson Starn, Kelermorph — ward {2}"),
			effects.On(game.EventDealDamage, effects.GapB194AnotherSourceYouControlDealtExactlyOne,
				"Ghyrson Starn, Kelermorph — he deals 2 damage to that permanent or player", effects.GapB194DealTwoToTheDamaged),
		},
	})
}
