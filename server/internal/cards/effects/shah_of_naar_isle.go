package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shah of Naar Isle — Creature — Efreet, {3}{R}, 6/6:
//
//	"Trample
//	 Echo {0} (At the beginning of your upkeep, if this came under your
//	 control since the beginning of your last upkeep, sacrifice it unless
//	 you pay its echo cost.)
//	 When this creature's echo cost is paid, each opponent may draw up to
//	 three cards."
//
// The echo is ADR 0108 §5's, and paying it emits EventEchoPaid, which the
// last line watches (WhenEchoIsPaid).
//
// "Each opponent may draw up to three cards" is a choice each opponent
// makes as the trigger resolves: none, one, two or three. The choices are
// made in turn order starting after Shah's controller (CR 101.4 — the
// controller is the active player, since echo is paid in their upkeep),
// and then everyone draws what they chose.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "31c978b4-06d7-4e9d-82da-bfa4ef16f470",
		Name:            "Shah of Naar Isle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Echo("Shah of Naar Isle", "{0}"),
			WhenEchoIsPaid("Shah of Naar Isle — each opponent may draw up to three cards", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return shahAskOpponents(ctx, opponentsInTurnOrderAfter(ctx, item.Controller), nil)
			}),
		},
	})
}

// shahDrawOptions are the answers to "you may draw up to three cards".
var shahDrawOptions = []game.ChoiceOption{
	{Label: "Draw no cards"},
	{Label: "Draw one card"},
	{Label: "Draw two cards"},
	{Label: "Draw three cards"},
}

// shahAskOpponents asks each of `opps` in order how many cards to draw,
// then draws them all. `chosen` holds the answers so far; it is copied,
// never shared, so an undo across one of the prompts replays cleanly.
func shahAskOpponents(ctx *Context, opps []uuid.UUID, chosen []int) error {
	if len(chosen) == len(opps) {
		for i, p := range opps {
			if err := (DrawCards{Player: p, N: chosen[i]}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
	// CR 121.2b: a draw limit caps how many the opponent may choose to
	// draw, and one who can't draw at all isn't asked (#4016).
	who := opps[len(chosen)]
	ceiling := ctx.Game.DrawChoiceCeilingForEffect(who, len(shahDrawOptions)-1)
	if ceiling <= 0 {
		return shahAskOpponents(ctx, opps, append(append([]int(nil), chosen...), 0))
	}
	return PickOption{
		Player:   who,
		Question: "Shah of Naar Isle — you may draw up to three cards",
		Options:  shahDrawOptions[:ceiling+1],
		ThenKey:  shahAnswer,
		Carry:    packIDLists(opps, intsToCarry(chosen)),
	}.Apply(ctx)
}

// shahAnswer is one opponent's answer, keyed (#4253 sweep): the opponents in
// order and the answers so far ride the prompt as IDs (intsToCarry).
// Registered in init, because the body asks the next opponent with it.
var shahAnswer game.OptionPickThen

func init() {
	shahAnswer = OptionPickThen("option-pick/shah-of-naar-isle-draw", func(ctx *Context, r game.OptionPicked) error {
		opps, answers := unpackIDPair(r.Carry)
		index := r.Index
		if index < 0 {
			index = 0
		}
		next := append(carryToInts(answers), index)
		return shahAskOpponents(ctx, opps, next)
	})
}

// opponentsInTurnOrderAfter is `player`'s opponents in turn order,
// starting with the one after them (CR 101.4).
func opponentsInTurnOrderAfter(ctx *Context, player uuid.UUID) []uuid.UUID {
	seats := ctx.Game.Seats
	start := 0
	for i, p := range seats {
		if p != nil && p.ID == player {
			start = i
			break
		}
	}
	var out []uuid.UUID
	for k := 1; k < len(seats); k++ {
		p := seats[(start+k)%len(seats)]
		if p == nil || p.Eliminated || p.ID == player {
			continue
		}
		out = append(out, p.ID)
	}
	return out
}
