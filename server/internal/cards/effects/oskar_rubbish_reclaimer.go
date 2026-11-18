package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oskar, Rubbish Reclaimer — Legendary Creature — Human Wizard {3}{U}{B}, 3/3
// (Streets of New Capenna):
//
//	"This spell costs {1} less to cast for each different mana value among
//	 cards in your graveyard.
//	 Whenever you discard a nonland card, you may cast it from your graveyard."
//
// The reduction is Oskar's OWN cost clause (SelfCostModifiers, CR 601.2f) and
// spends against the generic part only; the count is the caster's graveyard
// with the spell itself left out.
//
// The trigger fires once per nonland card discarded. "You may cast it" is a
// cast during the trigger's resolution (CR 608.2g), paying the card's mana
// cost: a graveyard cast permission over that one card, any timing, offered as
// a resolution_cast prompt (Conduit of Worlds' shape). Declining, or a card that
// has left the graveyard first (CR 400.7), leaves it where it is. A card you do
// not own is not in "your graveyard", so it is never offered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "14c2c818-0ce2-4809-a5ff-ee4a6253defa",
		Name:         "Oskar, Rubbish Reclaimer",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(sncCostLessPerDifferentMV,
				"This spell costs {1} less to cast for each different mana value among cards in your graveyard."),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return discardedByYou(ev, source) && conspiracyTheoristNonland(g, ev.CardID)
			}, "Oskar, Rubbish Reclaimer — you may cast the discarded card from your graveyard", oskarCastDiscarded),
		},
	})
}

// oskarCastDiscarded grants and offers the cast of the discarded card.
func oskarCastDiscarded(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.CardID == uuid.Nil {
		return nil
	}
	id, you := item.Trigger.Event.CardID, item.Controller
	z := g.FindCardZoneForEffect(id)
	c, ok := g.LookupCardForEffect(id)
	if !ok || z == nil || z.Kind != game.ZoneGraveyard || c.Owner != you || c.IsLand() {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:      you,
		Zone:        game.ZoneGraveyard,
		Duration:    g.UntilEndOfTurnDuration(),
		Timing:      game.TimingFlash,
		CastOnly:    true,
		LapseOnPass: game.LapseStaysInExile,
		Source:      item.SourceCardID,
		SourceName:  "Oskar, Rubbish Reclaimer",
		Label:       "Oskar, Rubbish Reclaimer — cast " + c.Name + " from your graveyard",
	}, []game.Card{c})
	offerResolutionCastOver(g, you, item.SourceCardID, "Oskar, Rubbish Reclaimer",
		game.MayCastKeywordPaidCast, declineLeaveInGraveyard, id)
	return nil
}
