package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion of Kamigawa // Rooftop Saboteurs — March of the Machine (#5477).
//
// Front, Battle — Siege 4, {3}{U}: "When this Siege enters, tap target
// artifact or creature an opponent controls and put a stun counter on it."
// Back, Rooftop Saboteurs, Creature — Moonfolk Ninja 2/3: "Flying. Whenever
// this creature deals combat damage to a player or battle, draw a card."
//
// No simplification.
const invasionOfKamigawaOracleID = "462fa2a4-651b-4fc1-b73b-1bd4508c8f75"

func init() {
	Register(Spec{
		OracleID:     invasionOfKamigawaOracleID,
		Name:         "Invasion of Kamigawa",
		Completeness: CompletenessFull,
		Battle:       &BattleSpec{Defense: 4, Subtype: BattleSubtypeSiege},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Invasion of Kamigawa — tap target artifact or creature an opponent controls and put a stun counter on it", tapAndStunTheTarget),
				TargetPermanent("target artifact or creature an opponent controls", Or(Artifact(), Creature()), OpponentControls())),
			DefeatedTrigger("Invasion of Kamigawa — defeated: exile it, then cast Rooftop Saboteurs", SiegeDefeated()),
		},
	})
	Register(Spec{
		OracleID:        invasionOfKamigawaOracleID + "#1",
		Name:            "Rooftop Saboteurs",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, thisDealtCombatDamageToAPlayerOrBattle,
				"Rooftop Saboteurs — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
