package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Twilight Shepherd — Creature — Angel {3}{W}{W}{W}, 5/5:
//
//	"Flying, vigilance
//	 When this creature enters, return to your hand all cards in your
//	 graveyard that were put there from the battlefield this turn.
//	 Persist"
//
// The cards are found as the trigger resolves, by their newest arrival
// in your graveyard (a card that left and came back is a new object,
// CR 400.7). When persist returns the Shepherd after a wipe, its own
// card has already left the graveyard, and everything else that died
// this turn comes back to your hand. Flying, vigilance and persist are
// PrintedKeywords (#2075).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eff736bc-d311-4bef-9c7e-de4ebe32e46f",
		Name:            "Twilight Shepherd",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", game.KeywordPersist},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Twilight Shepherd — return to your hand the cards put into your graveyard from the battlefield this turn",
				func(g *game.Game, item *game.StackItem) error {
					p := g.PlayerByIDForEffect(item.Controller)
					if p == nil || p.Graveyard == nil {
						return nil
					}
					var ids []uuid.UUID
					for _, c := range p.Graveyard.Cards {
						if putIntoGraveyardFromBattlefieldThisTurn(g, c.InstanceID) {
							ids = append(ids, c.InstanceID)
						}
					}
					for _, id := range ids {
						if err := g.ReturnFromGraveyardForEffect(id, game.ZoneHand); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
