package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Garna, the Bloodflame — Legendary Creature — Human Warrior {3}{B}{R}, 3/3
// (Commander deck reprint):
//
//	"Flash
//	 When Garna enters, return to your hand all creature cards in your
//	 graveyard that were put there from anywhere this turn.
//	 Other creatures you control have haste."
//
// The enter trigger reads Card.PutInGraveyardThisTurn, the per-turn arrival
// flag every zone move stamps on a card landing in a graveyard (cleared when
// it leaves and at the next turn), so a creature that died, was discarded or
// was milled this turn comes back and one that arrived on an earlier turn does
// not. The set is read as the trigger resolves. The haste is a layer 6 grant
// to the controller's other creatures (Aggressive Mammoth's shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "97cb993c-90ae-49ff-8ccb-b4fe317a43ef",
		Name:            "Garna, the Bloodflame",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Others: true, YoursOnly: true}, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Garna — return all creature cards put into your graveyard this turn to your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					p := g.PlayerByIDForEffect(ctx.Controller())
					if p == nil || p.Graveyard == nil {
						return nil
					}
					var ids []uuid.UUID
					for _, c := range p.Graveyard.Cards {
						if c.IsCreature() && c.PutInGraveyardThisTurn {
							ids = append(ids, c.InstanceID)
						}
					}
					for _, id := range ids {
						if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
