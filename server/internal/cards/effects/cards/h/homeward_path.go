package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Homeward Path — Land (Tales of Middle-earth Commander):
//
//	"{T}: Add {C}.
//	 {T}: Each player gains control of all creatures they own."
//
// Trostani Discordant's regaining of creatures, as an ability of the land:
// every creature on the battlefield whose controller is not its owner goes
// back to its owner, indefinitely (CR 611.2a), so a "for as long as you
// control" effect that gave it away has nothing left to restore; each is
// moved as a member of the group the ability read at resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "cb8ec2e4-8223-4172-8f2c-37c918a573fa",
		Name:          "Homeward Path",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.TapForColorless()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Each player gains control of all creatures they own.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect:  effects.Do(effects.Instruction(effects.BnsB22EachPlayerRegainsTheirCreatures("Homeward Path — its owner gains control of it"))),
		}},
	})
}
