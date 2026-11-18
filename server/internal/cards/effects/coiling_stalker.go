package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coiling Stalker — Creature — Snake Ninja {1}{G}, 2/1 (Neon Dynasty):
//
//	"Ninjutsu {1}{G}
//	 Whenever this creature deals combat damage to a player, put a +1/+1
//	 counter on target creature you control that doesn't have a +1/+1
//	 counter on it."
//
// The target is chosen as the trigger goes on the stack and re-checked as it
// resolves (CR 608.2b), so a creature that gained a counter in response is
// no longer a legal target and gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "54dc1c02-2d71-465b-a285-d6d11113991c",
		Name:         "Coiling Stalker",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Ninjutsu("{1}{G}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisDealsCombatDamageToAPlayer("Coiling Stalker — put a +1/+1 counter on target creature you control that doesn't have a +1/+1 counter on it",
				PlusOneCounterOnTheTarget),
				TargetCreature("target creature you control that doesn't have a +1/+1 counter on it", YouControl(), b13WithoutPlusOneCounter())),
		},
	})
}
