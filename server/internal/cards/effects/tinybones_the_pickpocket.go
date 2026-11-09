package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tinybones, the Pickpocket — Legendary Creature — Skeleton Rogue {B}, 1/1
// (Outlaws of Thunder Junction, #4468):
//
//	"Deathtouch
//	 Whenever Tinybones deals combat damage to a player, you may cast target
//	 nonland permanent card from that player's graveyard, and mana of any
//	 type can be spent to cast that spell."
//
// The clause names the player the damage was dealt to (Shark Shredder's
// shape). The cast happens as the trigger resolves (CR 608.2g): the card is
// granted a cast permission from the graveyard (mana of any type, CR
// 106.1b; the card's own mana cost is paid in full) and offered at once, a
// "no" leaving it in the graveyard. A card put somewhere else in response
// is not offered.
//
// No simplification.
func init() {
	damage := WheneverThisDealsCombatDamageToAPlayer(
		"Tinybones, the Pickpocket — you may cast target nonland permanent card from that player's graveyard, and mana of any type can be spent to cast that spell",
		tinybonesMayCast)
	damage.TargetsFrom = func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
		return TargetCardInGraveyard("target nonland permanent card from that player's graveyard",
			NonlandPermanentCard(), OwnedBy(tc.Event.Target))
	}
	Register(Spec{
		OracleID:        "7bc4c7e2-6758-4a85-84e7-03ab93981106",
		Name:            "Tinybones, the Pickpocket",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered:       []game.TriggeredAbility{damage},
	})
}

// tinybonesMayCast grants and offers the cast of the still-legal target.
func tinybonesMayCast(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	z := g.FindCardZoneForEffect(id)
	c, found := g.LookupCardForEffect(id)
	if !found || z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:      ctx.Controller(),
		Zone:        game.ZoneGraveyard,
		Duration:    g.UntilEndOfTurnDuration(),
		CastOnly:    true,
		Timing:      game.TimingFlash,
		AnyColor:    true,
		AnyType:     true,
		LapseOnPass: game.LapseStaysInHand,
		Source:      ctx.Source(),
		SourceName:  "Tinybones, the Pickpocket",
		Label:       "Tinybones, the Pickpocket — cast " + c.Name + " from that player's graveyard; mana of any type can be spent to cast it",
	}, []game.Card{c})
	offerResolutionCastOver(g, ctx.Controller(), ctx.Source(), "Tinybones, the Pickpocket", game.MayCastKeywordPaidCast, declineLeaveInGraveyard, c.InstanceID)
	return nil
}
