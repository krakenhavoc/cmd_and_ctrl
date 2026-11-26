package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Attendant of Vraska — Creature — Zombie Soldier {1}{B}{G}, 3/3 (Guilds of
// Ravnica, #5525):
//
//	"When this creature dies, if you control a Vraska planeswalker, you gain
//	 life equal to this creature's power."
//
// Bottle Golems' life gain (the dead creature's last-known power, CR 608.2h)
// behind an intervening "if" (CR 603.4): it is read as the creature dies and
// again as the trigger resolves, and the trigger does nothing if you no longer
// control a Vraska planeswalker.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9a639f82-24b6-4b54-8b9e-1fbc0f0094bf",
		Name:         "Attendant of Vraska",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you control a Vraska planeswalker", controlsAVraskaPlaneswalker,
				effects.WhenThisDies("Attendant of Vraska — you gain life equal to its power",
					effects.Do(effects.GainLifeEqualToSubject{Stat: effects.SubjectPower}))),
		},
	})
}

// controlsAVraskaPlaneswalker is "you control a Vraska planeswalker".
func controlsAVraskaPlaneswalker(g *game.Game, you uuid.UUID) bool {
	return effects.B40ControlsA(g, you, effects.And(effects.Planeswalker(), effects.HasSubtype("Vraska")))
}
