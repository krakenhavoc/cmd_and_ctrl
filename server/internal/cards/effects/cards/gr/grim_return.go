package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grim Return — Instant {2}{B} (Magic 2014, S82 OLD batch 401, #5563):
//
//	"Choose target creature card in a graveyard that was put there from the
//	 battlefield this turn. Put that card onto the battlefield under your
//	 control."
//
// Beacon of Unrest's reanimation over any graveyard, with the target narrowed by
// Faith's Reward's reading of "put there from the battlefield this turn": the
// card's newest arrival in its graveyard happened this turn and came from the
// battlefield (a card that was discarded, milled or countered is not one, and a
// token is not a card). The clause is judged again as the spell resolves
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c1d54d0c-53fc-474c-ba89-2aa1be3bf27d",
		Name:         "Grim Return",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("target creature card in a graveyard that was put there from the battlefield this turn",
			effects.Creature(), grimReturnPutThereThisTurn),
		Purpose: effects.ForTargets(effects.Reanimates(0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			effects.ReanimateSingleTarget(ctx, ctx.Controller())
			return nil
		},
	})
}

// grimReturnPutThereThisTurn is "that was put there from the battlefield this
// turn".
func grimReturnPutThereThisTurn(g *game.Game, _ uuid.UUID, c game.Card) bool {
	return effects.PutIntoGraveyardFromBattlefieldThisTurn(g, c.InstanceID)
}
