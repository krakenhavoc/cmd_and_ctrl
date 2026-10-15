package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Barret Wallace — Legendary Creature — Human Rebel {3}{R}, 4/4 (#3656):
//
//	"Reach
//	 Whenever Barret Wallace attacks, it deals damage equal to the
//	 number of equipped creatures you control to defending player."
//
// The count and the defending player are read as the trigger resolves
// (the player from the attack event, so it survives Barret leaving). A
// zero count deals no damage. Barret is the damage source, so lifelink
// or protection on him would apply.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "913672bc-cc88-4964-a331-7dbce40f2bda",
		Name:            "Barret Wallace",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Barret Wallace — deals damage equal to the number of equipped creatures you control to defending player",
				func(g *game.Game, item *game.StackItem) error {
					defender := g.DefendingPlayerForAttackForEffect(item.Trigger.Event.Target)
					if defender == uuid.Nil || g.PlayerByIDForEffect(defender) == nil {
						return nil
					}
					n := equippedCreaturesYouControl(g, item.Controller, false)
					return DealDamage{Source: item.SourceCardID, Target: defender, Amount: n}.Apply(NewContext(g, item))
				}),
		},
	})
}
