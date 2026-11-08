package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Maskmaker — Creature — Goblin Citizen {R}, 1/2 (Murders at
// Karlov Manor, #5241):
//
//	"Whenever this creature attacks, face-down spells you cast this turn
//	 cost {1} less to cast."
//
// The trigger leaves a promise on its controller for the rest of the
// turn (GrantNextSpellPromise with EveryMatchingSpell, Urza's and Armor
// Wars's shape), narrowed by NextSpellPromise.FaceDownOnly to a spell
// cast face down (CR 708.4): a disguise or morph cast's {3} costs {2}.
// It is a CR 601.2f reduction of generic mana, so the engine, the
// auto-tap preview and the legal-move enumerator all price it, and it
// outlives the Goblin. Two attacks in one turn make two promises, as
// the trigger resolving twice does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ff4b9723-d5f1-40d2-aa35-fa3d2a6f4b5c",
		Name:         "Goblin Maskmaker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Goblin Maskmaker — face-down spells you cast this turn cost {1} less to cast",
				Do(GrantNextSpellPromise{From: "Goblin Maskmaker", Promise: game.NextSpellPromise{
					FaceDownOnly:       true,
					Reduce:             1,
					EveryMatchingSpell: true,
					Text:               "Face-down spells you cast this turn cost {1} less to cast.",
				}})),
		},
	})
}
