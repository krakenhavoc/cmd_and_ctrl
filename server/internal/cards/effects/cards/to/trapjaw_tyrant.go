package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Trapjaw Tyrant — Creature — Dinosaur {3}{W}{W}, 5/5 (Rivals of Ixalan, S82
// OLD batch 218, #5531):
//
//	"Enrage — Whenever this creature is dealt damage, exile target creature an
//	 opponent controls until this creature leaves the battlefield."
//
// Boros Reckoner's enrage shape (one trigger per damage batch, CR 120.3) with
// Banisher Priest's exile: the target is chosen as the trigger goes on the
// stack and re-checked as it resolves (CR 608.2b); the card returns under its
// owner's control when the Tyrant leaves (CR 610.3c), and if the Tyrant has
// already left when the trigger resolves nothing is exiled (CR 610.3b). Lethal
// damage still triggers it, so a Tyrant that dies to the damage exiles nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00605b83-faf6-4779-be9d-6a2c8f0ea722",
		Name:         "Trapjaw Tyrant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Cmdr2B08EnrageTrigger("Trapjaw Tyrant — Enrage: exile target creature an opponent controls until this creature leaves the battlefield",
				effects.ExileChosenTargetUntilThisLeaves("Trapjaw Tyrant — the exiled creature returns when Trapjaw Tyrant leaves the battlefield")),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
