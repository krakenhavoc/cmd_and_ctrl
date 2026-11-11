package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Innocent Bystander — Creature — Goblin Citizen {1}{R}, 2/1 (Murders at
// Karlov Manor, #4469):
//
//	"Whenever this creature is dealt 3 or more damage, investigate.
//	 (Create a Clue token. It's an artifact with "{2}, Sacrifice this
//	 token: Draw a card.")"
//
// Barbed Servitor's "is dealt damage" trigger (b35SelfWasDealtDamage),
// settled at the end of the damage batch so the amount is the TOTAL
// dealt to it at once (CR 120.3, CR 510.2): two blockers dealing 2 each
// are one 4. Damage dealt at different times is never added up (the
// ruling). The settle runs before state-based actions, so a Bystander
// that the damage kills still investigates.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c293f1c0-93de-4a73-aaf6-4e9f01e557ec",
		Name:         "Innocent Bystander",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b35SelfWasDealtDamage(ev, source)
			},
			Key:        "Innocent Bystander — investigate (dealt 3 or more damage)",
			AtBatchEnd: &game.BatchEndCondition{Total: game.BatchTotalDamage, AtLeast: 3},
			Effect:     Do(Investigate{N: 1}),
		}},
	})
}
