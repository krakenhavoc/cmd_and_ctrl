package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Search Warrant — Sorcery {W}{U} (Return to Ravnica, #5566):
//
//	"Target player reveals their hand. You gain life equal to the number of cards
//	 in that player's hand."
//
// The hand is revealed to the table, then counted: the player is the target as it
// resolves, so a target that has become illegal (CR 608.2b) makes the whole spell
// do nothing. An empty hand reveals nothing and gains 0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27744033-5cb6-4d26-ad76-f4b4f1c0466c",
		Name:         "Search Warrant",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			p := ctx.Game.PlayerByIDForEffect(player)
			if p == nil || p.Hand == nil {
				return nil
			}
			ids := make([]uuid.UUID, 0, len(p.Hand.Cards))
			for _, c := range p.Hand.Cards {
				ids = append(ids, c.InstanceID)
			}
			if err := (effects.RevealCards{Player: player, Cards: ids, Reason: "Search Warrant — reveal their hand"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Amount: len(ids)}.Apply(ctx)
		},
	})
}
