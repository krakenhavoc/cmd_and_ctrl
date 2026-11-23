package cardsb

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bane, Lord of Darkness — {1}{W}{U}{B} Legendary Creature — God 5/2
//
//	"As long as your life total is less than or equal to half your starting
//	 life total, Bane has indestructible.
//	 Whenever another nontoken creature you control dies, target opponent may
//	 have you draw a card. If they don't, you may put a creature card with
//	 equal or lesser toughness from your hand onto the battlefield."
//
// The opponent is targeted as the trigger goes on the stack and answers as it
// resolves; a draw a limit stops (CR 121.2b) can't be chosen, so it counts as
// a "no". "Equal or lesser toughness" is the dead creature's toughness as it
// last existed (CR 608.2h), carried to the answer; the put is optional and
// enters the card without casting it.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ca507eea-1b04-4488-813e-ca970640e227",
		Name:         "Bane, Lord of Darkness",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.GapB214SelfIndestructibleAtHalfLife("Bane")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventLTB, effects.HofriGhostforgeAnotherNontokenCreatureDied,
					"Bane, Lord of Darkness — target opponent may have you draw a card; if they don't, you may put a creature card with equal or lesser toughness from your hand onto the battlefield",
					baneAsk),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}

func baneAsk(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	info, ok := ctx.TriggeringPermanent()
	if !ok {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		return effects.MayChoice{
			Player:   t.ID,
			Question: "Bane, Lord of Darkness — let its controller draw a card?",
			Draws:    1, Drawer: ctx.Controller(),
			YesLabel: "They draw a card",
			NoLabel:  "They may put a creature card from their hand onto the battlefield",
			Then:     baneAnswer,
			Word:     strconv.Itoa(info.Toughness),
		}.Apply(ctx)
	}
	return nil
}

// baneAnswer is the opponent's answer: yes draws, no (or an unavailable draw)
// lets Bane's controller put a creature card with toughness at most the dead
// creature's from their hand onto the battlefield. On-disk key.
var baneAnswer = effects.MayYesNo("confirm/bane-lord-of-darkness-draw-or-put",
	func(ctx *effects.Context, _ game.ConfirmAnswer) error {
		return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
	},
	func(ctx *effects.Context, r game.ConfirmAnswer) error {
		toughness, err := strconv.Atoi(r.Word)
		if err != nil {
			return nil
		}
		return effects.PutFromHandOntoBattlefield{
			Player:   ctx.Controller(),
			Match:    effects.And(effects.Creature(), effects.ToughnessLE(toughness)),
			Optional: true,
			Label:    "Bane, Lord of Darkness — you may put a creature card with toughness " + r.Word + " or less from your hand onto the battlefield",
		}.Apply(ctx)
	})
