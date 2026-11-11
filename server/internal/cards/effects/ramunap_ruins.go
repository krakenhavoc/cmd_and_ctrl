package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ramunap Ruins — Land — Desert:
//
//	"{T}: Add {C}.
//	 {T}, Pay 1 life: Add {R}.
//	 {2}{R}{R}, {T}, Sacrifice a Desert: This land deals 2 damage to each
//	 opponent."
//
// The Desert may be any one the controller has, this land included. The
// damage is dealt by this land (last-known information if it was the one
// sacrificed, CR 608.2h).
//
// No simplification.
func init() {
	Register(cmdr2B13DesertLand("d0d35864-1edc-4af1-9b89-3d7e94908011", "Ramunap Ruins", "R", ActivatedAbility{
		Label:   "{2}{R}{R}, {T}, Sacrifice a Desert: This land deals 2 damage to each opponent.",
		Purpose: game.Purpose{Answers: game.AnswerValue},
		Cost:    Plus(ManaCost("{2}{R}{R}"), TapCost(), cmdr2B13SacrificeADesert()),
		Effect:  Do(DamageEachOpponent{N: 2}),
	}))
}
