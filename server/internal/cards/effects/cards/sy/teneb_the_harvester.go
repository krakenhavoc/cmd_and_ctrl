package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Teneb, the Harvester — Legendary Creature — Dragon {3}{W}{B}{G}, 6/6
// (Alara Reborn):
//
//	"Flying
//	 Whenever Teneb deals combat damage to a player, you may pay {2}{B}.
//	 If you do, put target creature card from a graveyard onto the
//	 battlefield under your control."
//
// Flying is a printed keyword. The creature card is chosen as the trigger
// goes on the stack (CR 603.3d), before the payment is asked; the payment
// is the pay prompt as it resolves (CR 118.12), and only if the mana was
// really paid does the card come back under the controller's control, if
// it is still a legal pick in a graveyard (CR 608.2b). With no creature
// card in any graveyard the trigger is removed for want of a target.
//
// No simplification.
func init() {
	t := effects.WheneverThisDealsCombatDamageToAPlayer("Teneb, the Harvester — you may pay {2}{B}: put target creature card from a graveyard onto the battlefield under your control",
		func(g *game.Game, item *game.StackItem) error {
			return effects.MayPay{
				Chooser:  item.Controller,
				Cost:     "{2}{B}",
				Question: "Teneb, the Harvester — pay {2}{B} to put the target creature card onto the battlefield under your control?",
				Then:     tenebPaid,
			}.Apply(effects.NewContext(g, item))
		})
	t.Targets = effects.TargetCardInGraveyard("target creature card from a graveyard", effects.Creature())
	effects.Register(effects.Spec{
		OracleID:        "d41f0357-f1d6-45cc-b8ea-4efa41da1448",
		Name:            "Teneb, the Harvester",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{t},
	})
}

// tenebPaid is the "if you do". An on-disk key: never renamed, never
// reused.
var tenebPaid = effects.PayUnlessBranches("pay-unless/teneb-the-harvester-reanimate",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.GapB55PutTargetCreatureCardOntoTheBattlefield(ctx.Game, ctx.Item)
	}, nil)
