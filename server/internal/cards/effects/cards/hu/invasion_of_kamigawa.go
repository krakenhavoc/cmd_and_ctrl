package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Invasion of Kamigawa // Rooftop Saboteurs — March of the Machine (#5477).
//
// Front, Battle — Siege 4, {3}{U}: "When this Siege enters, tap target
// artifact or creature an opponent controls and put a stun counter on it."
// Back, Rooftop Saboteurs, Creature — Moonfolk Ninja 2/3: "Flying. Whenever
// this creature deals combat damage to a player or battle, draw a card."
//
// No simplification.
const InvasionOfKamigawaOracleID = "462fa2a4-651b-4fc1-b73b-1bd4508c8f75"

func init() {
	effects.Register(effects.Spec{
		OracleID:     InvasionOfKamigawaOracleID,
		Name:         "Invasion of Kamigawa",
		Completeness: effects.CompletenessFull,
		Battle:       &effects.BattleSpec{Defense: 4, Subtype: effects.BattleSubtypeSiege},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Invasion of Kamigawa — tap target artifact or creature an opponent controls and put a stun counter on it", effects.TapAndStunTheTarget),
				effects.TargetPermanent("target artifact or creature an opponent controls", effects.Or(effects.Artifact(), effects.Creature()), effects.OpponentControls())),
			effects.DefeatedTrigger("Invasion of Kamigawa — defeated: exile it, then cast Rooftop Saboteurs", effects.SiegeDefeated()),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        InvasionOfKamigawaOracleID + "#1",
		Name:            "Rooftop Saboteurs",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayerOrBattle,
				"Rooftop Saboteurs — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
