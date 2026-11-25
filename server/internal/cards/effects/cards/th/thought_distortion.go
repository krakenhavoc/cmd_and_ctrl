package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thought Distortion — Sorcery {4}{B}{B} (Core Set 2020):
//
//	"This spell can't be countered.
//	 Target opponent reveals their hand. Exile all noncreature, nonland cards
//	 from that player's hand and graveyard."
//
// The hand is revealed to every player (CR 701.20a), then every noncreature,
// nonland card in that hand and in that graveyard leaves for exile as one
// simultaneous move (CR 400.7). Nothing is chosen. The spell's "can't be
// countered" is the spec's flag (CR 101.2). An illegal target at resolution
// does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5f089ac6-9e92-4ec2-bf46-a0b08d1e2979",
		Name:            "Thought Distortion",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		Targets:         effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			who := effects.TargetedPlayer(ctx)
			p := ctx.Game.PlayerByIDForEffect(who)
			if p == nil || p.Hand == nil {
				return nil
			}
			ctx.Game.RevealHandForEffect(who)
			match := effects.And(effects.Noncreature(), effects.Nonland())
			var ids []uuid.UUID
			for _, z := range []*game.Zone{p.Hand, p.Graveyard} {
				if z == nil {
					continue
				}
				for _, c := range z.Cards {
					if match(ctx.Game, ctx.Controller(), c) {
						ids = append(ids, c.InstanceID)
					}
				}
			}
			ctx.Game.ExileCardsForEffect(ids)
			return nil
		},
	})
}
