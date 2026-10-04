package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// It That Betrays — Creature — Eldrazi {12}, 11/11:
//
//	"Annihilator 2
//	 Whenever an opponent sacrifices a nontoken permanent, put that card
//	 onto the battlefield under your control."
//
// Annihilator 2 is the canonical keyword token (ADR 0113 §2), and it is
// the obvious way to feed the second ability — but the 2010-06-15
// rulings say the second triggers on ANY sacrifice by an opponent, to
// annihilator, an edict or a cost, and that it does not matter whose
// graveyard the card went to. The sacrifice is announced while the
// permanent is still on the battlefield, so "nontoken" is read there.
// The card is put onto the battlefield only if it is still in a
// graveyard as the trigger resolves; one that has left (exiled, or
// taken by another It That Betrays) stays where it is.
func init() {
	Register(Spec{
		OracleID:        "b11187ab-f8a8-422b-b550-4495f96de0f2",
		Name:            "It That Betrays",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"An Aura returned this way comes back unattached and is put into the graveyard — you don't get to choose what it enchants."},
		PrintedKeywords: []string{"annihilator 2"},
		Triggered: []game.TriggeredAbility{
			On(game.EventSacrifice, AnOpponentSacrificedANontokenPermanent,
				"It That Betrays — put that card onto the battlefield under your control",
				putTheSacrificedCardOntoTheBattlefieldUnderYourControl),
		},
	})
}

// AnOpponentSacrificedANontokenPermanent — "whenever an opponent
// sacrifices a nontoken permanent". EventSacrifice is emitted while the
// permanent is still on the battlefield, with the sacrificing player in
// Actor, so the card is looked up there.
func AnOpponentSacrificedANontokenPermanent(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventSacrifice || ev.CardID == uuid.Nil || ev.Actor == uuid.Nil || ev.Actor == source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.IsToken()
}

// putTheSacrificedCardOntoTheBattlefieldUnderYourControl puts the card
// the triggering sacrifice named onto the battlefield under the
// trigger's controller, if it is still in a graveyard.
func putTheSacrificedCardOntoTheBattlefieldUnderYourControl(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id := ctx.Trigger().Event.CardID
	if z := g.FindCardZoneForEffect(id); id == uuid.Nil || z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: item.Controller}.Apply(ctx)
}
