package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Done for the Day — Enchantment {2}{G}:
//
//	"At the beginning of your end step, if you control an Employee, a
//	 Performer, or a Robot, you may get {TK} or create a Treasure token. If
//	 you control all three, you may put a sticker on a nonland permanent you
//	 own."
//
// The intervening "if" (CR 603.4) is checked as the end step begins and
// again as the ability resolves; each step of the one resolution checks it
// for itself:
//
//  1. "You may get {TK} or create a Treasure token": a keyed option pick of
//     "neither", "get {TK}" and "create a Treasure token" (a "may" with a
//     choice is one question).
//  2. "If you control all three" is read when the step begins, after the
//     first has settled (CR 608.2c): an Employee, a Performer and a Robot
//     among permanents you control (a creature with all three types, or a
//     changeling, counts for each). Then the sticker is the engine's keyed
//     questions (ADR 0208 §4), "you may" like every sticker.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2a248d6a-aa83-4e59-848b-46de8204b538",
		Name:         "Done for the Day",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Actor == source.Controller && controlsAnyOfEmployeePerformerRobot(g, source.Controller, source.TextSubtype)
				},
				"Done for the Day — you may get {TK} or create a Treasure token. If you control an Employee, a Performer and a Robot, you may put a sticker on a nonland permanent you own",
				Do(
					Instruction(doneForTheDayReward),
					Instruction(doneForTheDaySticker),
				)),
		},
	})
}

// controlledEmployeePerformerRobot says which of the three `you` control.
func controlledEmployeePerformerRobot(g *game.Game, you uuid.UUID, tr func(string) string) (employee, performer, robot bool) {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != you {
			continue
		}
		employee = employee || c.HasSubtype(tr("Employee"))
		performer = performer || c.HasSubtype(tr("Performer"))
		robot = robot || c.HasSubtype(tr("Robot"))
	}
	return employee, performer, robot
}

func controlsAnyOfEmployeePerformerRobot(g *game.Game, you uuid.UUID, tr func(string) string) bool {
	e, p, r := controlledEmployeePerformerRobot(g, you, tr)
	return e || p || r
}

// doneForTheDayRewardPick is the answer to "get {TK} or create a Treasure
// token": 1 is the ticket, 2 the Treasure, 0 neither.
var doneForTheDayRewardPick = OptionPickThen("option-pick/done-for-the-day-reward", func(ctx *Context, r game.OptionPicked) error {
	switch r.Index {
	case 1:
		return GetTickets{N: 1}.Apply(ctx)
	case 2:
		return CreateToken{Template: TreasureToken(), N: 1}.Apply(ctx)
	}
	return nil
})

func doneForTheDayReward(ctx *Context) error {
	if !controlsAnyOfEmployeePerformerRobot(ctx.Game, ctx.Controller(), ctx.TextSubtype) {
		return nil
	}
	return PickOption{
		Question: "Done for the Day — get {TK} or create a Treasure token?",
		Options: []game.ChoiceOption{
			{Label: "Neither"},
			{Label: "Get {TK}"},
			{Label: "Create a Treasure token"},
		},
		ThenKey: doneForTheDayRewardPick,
	}.Apply(ctx)
}

func doneForTheDaySticker(ctx *Context) error {
	e, p, r := controlledEmployeePerformerRobot(ctx.Game, ctx.Controller(), ctx.TextSubtype)
	if !(e && p && r) {
		return nil
	}
	return PutASticker{}.Apply(ctx)
}
