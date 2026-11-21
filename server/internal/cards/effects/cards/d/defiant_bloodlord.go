package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Defiant Bloodlord — Creature — Vampire {5}{B}{B}, 4/5 (Battle for
// Zendikar, #5547):
//
//	"Flying
//	 Whenever you gain life, target opponent loses that much life."
//
// Sanguine Bond's trigger on a flyer: one trigger per life-gain event,
// the amount read off the event, and the opponent is a target chosen as
// it goes on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7bf5196d-a16c-418f-943c-2128e40c55da",
		Name:            "Defiant Bloodlord",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouGainLifeTargetOpponentLosesThatMuch("Defiant Bloodlord — target opponent loses that much life"),
		},
	})
}
