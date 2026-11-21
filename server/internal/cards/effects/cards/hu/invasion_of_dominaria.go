package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Invasion of Dominaria // Serra Faithkeeper — March of the Machine (#5477).
//
// Front, Battle — Siege 5, {2}{W}: "When this Siege enters, you gain 4 life
// and draw a card." Back, Serra Faithkeeper, Creature — Angel 4/4: "Flying,
// vigilance". The protector prompt and defense counters are engine behaviour
// keyed on the battle type (game/battle.go).
//
// No simplification.
const InvasionOfDominariaOracleID = "dcbd8b15-e5fd-4f0d-ba2d-06f018b64fb3"

func init() {
	effects.Register(effects.Spec{
		OracleID:     InvasionOfDominariaOracleID,
		Name:         "Invasion of Dominaria",
		Completeness: effects.CompletenessFull,
		Battle:       &effects.BattleSpec{Defense: 5, Subtype: effects.BattleSubtypeSiege},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Invasion of Dominaria — you gain 4 life and draw a card", effects.Do(effects.GainLife{Amount: 4}, effects.DrawCards{N: 1})),
			effects.DefeatedTrigger("Invasion of Dominaria — defeated: exile it, then cast Serra Faithkeeper", effects.SiegeDefeated()),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        InvasionOfDominariaOracleID + "#1",
		Name:            "Serra Faithkeeper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
}
