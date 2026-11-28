package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Amulet of Quoz — Artifact {6} (Ice Age, S82 OLD #5636):
//
//	"Remove this card from your deck before playing if you're not playing for
//	 ante.
//	 {T}, Sacrifice this artifact: Target opponent may ante the top card of their
//	 library. If they don't, you flip a coin. If you win the flip, that player
//	 loses the game. If you lose the flip, you lose the game. Activate only during
//	 your upkeep."
//
// The ante clause is inert (docs/adding-cards.md, "Ante cards: built, with their
// ante text inert", #5724, CR 407): C2 has no ante zone, so no opponent antes
// and "if they don't" always holds; the coin flip, and the loss it can cost the
// Amulet's controller, are live. The deck-building sentence is not an ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a06f8476-0a23-413a-bc28-8dc2427fdc2a",
		Name:         "Amulet of Quoz",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}, Sacrifice this artifact: Target opponent may ante the top card of their library. If they don't, you flip a coin. If you win the flip, that player loses the game. If you lose the flip, you lose the game. Activate only during your upkeep.",
			Cost:      effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets:   effects.TargetPlayer("target opponent", effects.Opponent()),
			Condition: effects.DuringYourUpkeep(),
			Effect:    effects.AmuletOfQuozB832,
		}},
	})
}
