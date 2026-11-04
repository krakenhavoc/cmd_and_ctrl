package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Keen Duelist — Creature — Human Wizard {1}{B}, 2/2:
//
//	"At the beginning of your upkeep, you and target opponent each reveal
//	 the top card of your library. You each lose life equal to the mana
//	 value of the card revealed by the other player. You each put the card
//	 you revealed into your hand."
//
// Dark Confidant's flip for two. Both cards are revealed from the library
// first, each mana value is read before anything moves, then each player
// loses the other's number and each card goes to its owner's hand.
// Revealing is not drawing, so no draw trigger sees it, and an empty
// library reveals nothing and costs the other player nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b8602f19-6fc2-43c4-b0c4-2cd11d5dbda2",
		Name:         "Keen Duelist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(AtYourUpkeep("Keen Duelist — you and target opponent each reveal the top card of your library", keenDuelistFlip),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

// keenDuelistFlip is the whole trigger.
func keenDuelistFlip(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := firstLegalTarget(ctx)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	players := []uuid.UUID{item.Controller, t.ID}
	revealed := make([]uuid.UUID, len(players))
	values := make([]int, len(players))
	for i, p := range players {
		var ids []uuid.UUID
		if err := (RevealTopOfLibrary{
			Player:   p,
			N:        1,
			Reason:   "Keen Duelist — reveal the top card of your library",
			Revealed: &ids,
		}).Apply(ctx); err != nil {
			return err
		}
		if len(ids) == 0 {
			continue
		}
		revealed[i] = ids[0]
		if c, ok := g.LookupCardForEffect(ids[0]); ok {
			values[i] = c.ManaValue()
		}
	}
	for i, p := range players {
		if lose := values[1-i]; lose > 0 {
			if err := g.ChangePlayerLifeForEffect(item.SourceCardID, p, -lose); err != nil {
				return err
			}
		}
	}
	for _, id := range revealed {
		if id == uuid.Nil {
			continue
		}
		if err := (BounceToHand{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
