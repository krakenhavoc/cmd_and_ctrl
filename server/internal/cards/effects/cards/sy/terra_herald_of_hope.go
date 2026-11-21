package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Terra, Herald of Hope — Legendary Creature — Human Wizard Warrior
// {R}{W}{B}, 3/3 (deck request #3655):
//
//	"Trance — At the beginning of combat on your turn, mill two cards.
//	 Terra gains flying until end of turn.
//	 Whenever Terra deals combat damage to a player, you may pay {2}.
//	 When you do, return target creature card with power 3 or less from
//	 your graveyard to the battlefield tapped."
//
// The combat trigger mills first and grants flying second, in printed
// order. The damage trigger is the MayPay / CR 603.12 pair Rodolf
// Duskbringer uses: the payment is the whole condition, and the return
// is a second triggered ability whose target is chosen when it goes on
// the stack, so a creature card milled in response is a legal pick.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25d40d28-1b46-4055-959a-45719040ed30",
		Name:         "Terra, Herald of Hope",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Terra, Herald of Hope — mill two cards, gains flying until end of turn", terraTrance),
			effects.WheneverThisDealsCombatDamageToAPlayer("Terra, Herald of Hope — you may pay {2} to return a creature card with power 3 or less", terraPayOffer),
		},
	})
}

// terraTrance is the beginning-of-combat body: mill two, then flying.
func terraTrance(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.MillCards{Player: item.Controller, N: 2}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{
		Target:   item.SourceCardID,
		Keywords: []string{"flying"},
		Label:    "Terra, Herald of Hope — flying until end of turn",
	}.Apply(ctx)
}

// terraPayOffer asks for the {2}; the reflexive trigger lives in OnPay.
func terraPayOffer(g *game.Game, item *game.StackItem) error {
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{2}",
		Question: "Terra, Herald of Hope — pay {2} to return a creature card with power 3 or less from your graveyard to the battlefield tapped?",
		Then:     terraPaid,
	}.Apply(effects.NewContext(g, item))
}

// terraPaid is the paid half, keyed (#4264): the reflexive trigger is queued against the asking item.
var terraPaid = effects.PayUnlessBranches("pay-unless/terra-herald-of-hope-return",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.ReflexiveTrigger{
			Label: "Terra, Herald of Hope — return target creature card with power 3 or less to the battlefield tapped",
			Body:  effects.TerraReturnBody,
		}.Apply(ctx)
	}, nil)
