package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// City Hall — Plane — Chicago (punk, #4870):
//
//	"When you planeswalk here, you become the Monarch.
//	 Whenever Chaos ensues, each player may create two tapped treasure
//	 tokens. Each player who does can't attack you during their next
//	 turn."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6). The
// chaos ability (CR 311.7) asks each player in turn order from the active
// player whether they want the Treasures (a keyed option pick, so a table
// waiting on it is a restore point); each who says yes creates two tapped
// Treasure tokens and cannot attack the planar controller (the player, not
// their planeswalkers: the clause says "you") during their next turn, the
// restriction CR 508.1c allows.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "36aac8aa-d6f0-4e53-a280-73b01e040f35",
		Name:         "City Hall",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThis("City Hall — you become the monarch", Do(BecomeTheMonarch{})),
			WheneverChaosEnsues("City Hall — each player may create two tapped Treasure tokens; each who does can't attack you during their next turn",
				func(g *game.Game, item *game.StackItem) error {
					return cityHallAsk(NewContext(g, item), seatsFromActive(g))
				}),
		},
	})
}

// cityHallAsk asks the first of `players` still in the game.
func cityHallAsk(ctx *Context, players []uuid.UUID) error {
	for len(players) > 0 {
		p := ctx.PlayerByID(players[0])
		if p == nil || p.Eliminated {
			players = players[1:]
			continue
		}
		return PickOption{
			Player:   players[0],
			Question: "City Hall — create two tapped Treasure tokens? You can't attack the planar controller during your next turn.",
			Options: []game.ChoiceOption{
				{Label: "Create two tapped Treasure tokens"},
				{Label: "Don't"},
			},
			ThenKey:  cityHallAnswered,
			Carry:    append([]uuid.UUID(nil), players...),
			WithItem: true,
		}.Apply(ctx)
	}
	return nil
}

// cityHallAnswered applies one player's answer and asks the next. An on-disk
// key: never renamed, never reused. Registered in init: the answer asks the
// next question, which names this key.
var cityHallAnswered game.OptionPickThen

func init() { cityHallAnswered = OptionPickThen("option-pick/city-hall-treasures", cityHallAnswer) }

func cityHallAnswer(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || r.Option == nil {
		return nil
	}
	player := r.Carry[0]
	if r.Index == 0 {
		if err := (CreateTokenAdvanced{Controller: player, Spec: Token(TreasureToken()).EntersTapped(), N: 2}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
		ctx.Game.GrantCantAttackPlayerOnlyNextTurnForEffect(player, ctx.Controller(), "City Hall — can't attack the planar controller during their next turn", ctx.Item.SourceCardID)
	}
	return cityHallAsk(ctx, r.Carry[1:])
}
