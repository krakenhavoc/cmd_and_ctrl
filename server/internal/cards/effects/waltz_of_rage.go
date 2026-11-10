package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Waltz of Rage — Sorcery {3}{R}{R} (Duskmourn: House of Horror, #4466):
//
//	"Target creature you control deals damage equal to its power to each
//	 other creature. Until end of turn, whenever a creature you control
//	 dies, exile the top card of your library. You may play it until the end
//	 of your next turn."
//
// Two instructions in printed order (Nova Flame's shape: the second waits
// for any prompt the damage asks, CR 608.2c). The damage is the target's,
// not the spell's, one damage event to every other creature on the battlefield
// with the power read as the spell resolves (targetDealsItsPowerToEachOtherCreature).
// A target that is illegal as the spell resolves makes the whole spell fizzle
// (CR 608.2b): no damage and no delayed ability, as the ruling says.
//
// "Until end of turn, whenever a creature you control dies" is a CR 603.7c
// repeating delayed triggered ability (Electric Seaweed's shape): it exists
// from the moment the spell finishes its damage, so creatures of yours that
// die to that very damage (state-based actions run after the spell) trigger
// it, and each one exiles the top card of your library with the impulse
// window of Crimson Operative (until the end of your NEXT turn, so cast on
// your own turn it runs through the following one). Creatures that die in any
// player's turn this turn count, tokens included; a creature that dies as it
// is exiled instead does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "78df47c3-b771-4377-8963-ae3065fdcf8a",
		Name:         "Waltz of Rage",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: ResolveSteps(
			Instruction(targetDealsItsPowerToEachOtherCreature),
			DelayedOnEvent{
				Label:     "Waltz of Rage — exile the top card of your library; you may play it until the end of your next turn",
				On:        []game.EventKind{game.EventLTB},
				Condition: b24ACreatureYouControlDiedCondition,
				Body:      b24ExileTopYouMayPlayUntilNextTurnBody,
				Repeats:   true,
			},
		),
	})
}
