package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Embalmed Ascendant — Creature — Zombie {1}{W}{B}, 1/2:
//
//	"Start your engines!
//	 When this creature enters, create a 2/2 black Zombie creature token.
//	 Max speed — Whenever a creature you control dies, each opponent loses
//	 1 life and you gain 1 life."
//
// ADR 0138 (#2711). The drain counts the Ascendant's own death (it is a
// creature you control), and is gated on its controller's max speed as
// the creature dies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2f4d7988-3f5b-45d7-9838-02897d7180be",
		Name:            "Embalmed Ascendant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{effects.StartYourEngines},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Embalmed Ascendant — create a 2/2 black Zombie",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 1})),
			effects.MaxSpeedTrigger(effects.TriggerWithPurpose(
				effects.WheneverACreatureYouControlDies("Embalmed Ascendant — each opponent loses 1 life, you gain 1 life",
					effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})),
				game.Purpose{DeathPayoff: true})),
		},
	})
}
