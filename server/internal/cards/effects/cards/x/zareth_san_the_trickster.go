package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zareth San, the Trickster — Legendary Creature — Merfolk Rogue {3}{U}{B},
// 4/4 (Zendikar Rising, S82 OLD #5508):
//
//	"Flash
//	 {2}{U}{B}, Return an unblocked attacking Rogue you control to its
//	 owner's hand: Put this card from your hand onto the battlefield tapped
//	 and attacking.
//	 Whenever Zareth San deals combat damage to a player, you may put target
//	 permanent card from that player's graveyard onto the battlefield under
//	 your control."
//
// The hand ability is ninjutsu's body with a Rogue to return
// (S82OldB091ZarethSanAbility). The damage trigger targets a permanent card
// in the graveyard of the player that was dealt damage, chosen as the trigger
// goes on the stack, and puts it onto the battlefield under the controller's
// control if they choose to.
//
// No simplification.
func init() {
	damage := effects.WheneverThisDealsCombatDamageToAPlayer(
		"Zareth San — you may put target permanent card from that player's graveyard onto the battlefield under your control",
		func(g *game.Game, item *game.StackItem) error {
			return effects.MayChoice{
				Question: "Zareth San — put target permanent card from that player's graveyard onto the battlefield under your control?",
				YesLabel: "Put it onto the battlefield",
				NoLabel:  "Decline",
				Then:     zarethSanReanimate,
			}.Apply(effects.NewContext(g, item))
		})
	damage.TargetsFrom = func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
		return effects.TargetCardInGraveyard("target permanent card from that player's graveyard",
			effects.Permanent(), effects.OwnedBy(tc.Event.Target))
	}
	effects.Register(effects.Spec{
		OracleID:        "a316b0ce-1f63-4a8e-9a6e-bfbe03ca4e9e",
		Name:            "Zareth San, the Trickster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Activated:       []effects.ActivatedAbility{effects.S82OldB091ZarethSanAbility()},
		Triggered:       []game.TriggeredAbility{damage},
	})
}

var zarethSanReanimate = effects.MayYes("confirm/zareth-san-reanimate", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	return effects.S82Mom16PutTargetOntoTheBattlefield(ctx.Game, ctx.Item)
})
