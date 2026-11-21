package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Akki Battle Squad — Creature — Goblin Samurai {5}{R}, 6/6 (Kamigawa: Neon
// Dynasty):
//
//	"Whenever one or more modified creatures you control attack, untap all
//	 modified creatures you control. After this phase, there is an
//	 additional combat phase. This ability triggers only once each turn.
//	 (Equipment, Auras you control, and counters are modifications.)"
//
// A batched attack trigger (OncePerBatch: one trigger for the whole
// declaration) that TriggersOnlyOnceEachTurn spends the first time it goes
// on the stack, as Mirelurk Queen's does. It resolves as Aurelia's does:
// untap, then ExtraCombatAfterThisPhase. Which creatures are modified is read
// as it resolves (CR 700.9, b07IsModified); the extra combat is anchored to
// the phase it triggered in (the declare-attackers step's combat phase), and
// comes with no main phase (CR 500.8).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5bd81d2f-46a1-4536-92e7-28c7e1892735",
		Name:         "Akki Battle Squad",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggersOnlyOnceEachTurn(effects.OncePerBatch(effects.On(game.EventAttack,
				effects.S82NeoB11ModifiedCreatureYouControlAttacked,
				"Akki Battle Squad — untap all modified creatures you control; after this phase, there is an additional combat phase",
				effects.Do(effects.S82NeoB11UntapModified{}, effects.ExtraCombatAfterThisPhase())))),
		},
	})
}
