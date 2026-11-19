package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tethmos High Priest — Creature — Cat Cleric {2}{W}, 2/3 (Foundations Jumpstart):
//
//	"Heroic — Whenever you cast a spell that targets this creature, return
//	 target creature card with mana value 2 or less from your graveyard to
//	 the battlefield."
//
// Heroic is the shared cast trigger (CR 207.2c). The graveyard target is
// chosen as the trigger goes on the stack and re-checked as it resolves
// (CR 608.2b); with no such card the trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f922150c-8f79-412b-a611-59064e3e3cea",
		Name:         "Tethmos High Priest",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				heroic("Tethmos High Priest — return target creature card with mana value 2 or less from your graveyard to the battlefield (heroic)",
					returnFirstLegalGraveyardTargetToBattlefield),
				TargetCardInGraveyard("target creature card with mana value 2 or less in your graveyard",
					YouOwn(), Creature(), ManaValueLE(2))),
		},
	})
}
