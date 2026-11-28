package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jangling Automaton — Artifact Creature — Construct {3}, 3/2 (Weatherlight):
//
//	"Whenever this creature attacks, untap all creatures defending player
//	 controls."
//
// The defending player is the player attacked, or the controller of the
// planeswalker or battle attacked (CR 506.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6f67539-9f76-42b2-96a8-575f27ba928e",
		Name:         "Jangling Automaton",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Jangling Automaton — untap all creatures defending player controls",
				func(g *game.Game, item *game.StackItem) error {
					defender := effects.DefendingPlayerOfAttack(g, item.Trigger.Event)
					if defender == uuid.Nil {
						return nil
					}
					ctx := effects.NewContext(g, item)
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == defender && c.IsCreature() && c.Tapped {
							ids = append(ids, c.InstanceID)
						}
					}
					for _, id := range ids {
						if err := (effects.UntapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
