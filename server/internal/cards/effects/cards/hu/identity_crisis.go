package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Identity Crisis — Sorcery {2}{W}{W}{B}{B} (Alara Reborn):
//
//	"Exile all cards from target player's hand and graveyard."
//
// Both piles are read as the spell resolves and exiled together (CR 406.3);
// the cards are not revealed or chosen. A target player who is no longer
// legal makes the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0abe0584-d876-4827-972b-589da53dbe96",
		Name:         "Identity Crisis",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			victim := effects.FirstLegalTargetPlayer(ctx)
			p := ctx.PlayerByID(victim)
			if victim == uuid.Nil || p == nil {
				return nil
			}
			ids := make([]uuid.UUID, 0, len(p.Hand.Cards)+len(p.Graveyard.Cards))
			for _, c := range p.Hand.Cards {
				ids = append(ids, c.InstanceID)
			}
			for _, c := range p.Graveyard.Cards {
				ids = append(ids, c.InstanceID)
			}
			for _, id := range ids {
				if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
