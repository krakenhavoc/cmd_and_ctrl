package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion of Dominaria // Serra Faithkeeper — March of the Machine (#5477).
//
// Front, Battle — Siege 5, {2}{W}: "When this Siege enters, you gain 4 life
// and draw a card." Back, Serra Faithkeeper, Creature — Angel 4/4: "Flying,
// vigilance". The protector prompt and defense counters are engine behaviour
// keyed on the battle type (game/battle.go).
//
// No simplification.
const invasionOfDominariaOracleID = "dcbd8b15-e5fd-4f0d-ba2d-06f018b64fb3"

func init() {
	Register(Spec{
		OracleID:     invasionOfDominariaOracleID,
		Name:         "Invasion of Dominaria",
		Completeness: CompletenessFull,
		Battle:       &BattleSpec{Defense: 5, Subtype: BattleSubtypeSiege},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Invasion of Dominaria — you gain 4 life and draw a card", Do(GainLife{Amount: 4}, DrawCards{N: 1})),
			DefeatedTrigger("Invasion of Dominaria — defeated: exile it, then cast Serra Faithkeeper", SiegeDefeated()),
		},
	})
	Register(Spec{
		OracleID:        invasionOfDominariaOracleID + "#1",
		Name:            "Serra Faithkeeper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
}
