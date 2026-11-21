package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Second Doctor — Legendary Creature — Time Lord Doctor {2}{W}{U},
// 2/4:
//
//	"Players have no maximum hand size.
//	 How Civil of You — At the beginning of your end step, each player
//	 may draw a card. Each opponent who does can't attack you or
//	 permanents you control during their next turn."
//
// The hand-size line is Price of Knowledge's (HandSizeEachPlayer, ADR
// 0113 §3). The trigger asks every player in turn order starting with
// the Doctor's controller (CR 101.4: it is their end step, so they are
// the active player), each "may draw a card?", and only then does
// anyone draw, so no answer can depend on a card already drawn. Each
// player who said yes draws, and each of those who is an opponent of the
// controller is stopped from attacking the controller or the
// controller's permanents during THEIR next turn
// (game.GrantCantAttackPlayerForEffect, ADR 0063's amendment of
// 2026-10-08, #2109). The controller's own draw does nothing to them.
//
// "Who does" is "who said yes": a player with an empty library who
// agrees still draws nothing and is still restricted, which is the
// printed choice (they can decline).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8a6eda31-2791-4def-b9e7-3adde155a4f0",
		Name:         "The Second Doctor",
		Completeness: effects.CompletenessFull,
		HandSize:     []game.HandSizeStatic{effects.PlayersHaveNoMaxHandSize()},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("The Second Doctor — each player may draw a card", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				order := append([]uuid.UUID{item.Controller}, effects.OpponentsInTurnOrderAfter(ctx, item.Controller)...)
				return secondDoctorAsk(ctx, order, 0, nil)
			}),
		},
	})
}

// secondDoctorAsk asks order[i] whether to draw, then the next player;
// `drew` holds the players who said yes so far. It is copied, never
// shared, so an undo across one of the prompts replays cleanly.
func secondDoctorAsk(ctx *effects.Context, order []uuid.UUID, i int, drew []uuid.UUID) error {
	if i >= len(order) {
		return secondDoctorFinish(ctx, drew)
	}
	asked := order[i]
	return effects.MayChoice{
		Player:   asked,
		Question: "The Second Doctor — draw a card?",
		Draws:    1,
		Then:     SecondDoctorAnswer,
		Carry:    effects.JoinCarry(drew, order[i+1:]),
	}.Apply(ctx)
}

// secondDoctorFinish draws for everyone who agreed, then restricts the
// opponents among them.
func secondDoctorFinish(ctx *effects.Context, drew []uuid.UUID) error {
	controller := ctx.Controller()
	for _, p := range drew {
		if err := (effects.DrawCards{Player: p, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	for _, p := range drew {
		if p == controller {
			continue
		}
		ctx.Game.GrantCantAttackPlayerForEffect(p, controller,
			"The Second Doctor — can't attack its controller or their permanents", ctx.Source())
	}
	return nil
}

// secondDoctorAnswer is one player's answer to "draw a card?": a yes counts
// them among those who will draw, either way the next player is asked. The
// carried IDs are those who have said yes, a Nil, then the players still to
// ask. Registered in init because it asks the next player through
// secondDoctorAsk, which names it. Keyed (#4291).
var SecondDoctorAnswer game.ConfirmThen

func init() {
	SecondDoctorAnswer = effects.MayAnswer("confirm/the-second-doctor-draw", func(ctx *effects.Context, r game.ConfirmAnswer) error {
		drew, rest := effects.SplitCarry(r)
		if r.Accepted {
			drew = append(drew, r.Chooser)
		}
		return secondDoctorAsk(ctx, rest, 0, drew)
	})
}
