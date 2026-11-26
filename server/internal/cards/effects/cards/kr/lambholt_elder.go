package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lambholt Elder // Silverpelt Werewolf — Creature — Human Werewolf {2}{G}
// 1/2 // Creature — Werewolf 4/5:
//
//	Lambholt Elder: "At the beginning of each upkeep, if no spells were cast
//	 last turn, transform this creature."
//	Silverpelt Werewolf: "Whenever this creature deals combat damage to a
//	 player, draw a card.
//	 At the beginning of each upkeep, if a player cast two or more spells last
//	 turn, transform this creature."
//
// The two upkeep triggers are the Innistrad werewolf pair (CR 702.145's
// predecessor wording, intervening "if" CR 603.4).
func init() {
	const oracle = "12330f42-0e3f-40e1-86f3-57a34403d143"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Lambholt Elder",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.TransformAtUpkeepIfNoSpellsLastTurn("Lambholt Elder")},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Silverpelt Werewolf",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Silverpelt Werewolf — draw a card", effects.Do(effects.DrawCards{N: 1})),
			effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Silverpelt Werewolf"),
		},
	})
}
