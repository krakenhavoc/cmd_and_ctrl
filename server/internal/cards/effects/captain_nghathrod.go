package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// nghaMillLabel and nghaReanimateLabel are the stack labels (and so the
// row Keys) of Captain N'ghathrod's two triggers.
const (
	nghaMillLabel      = "Captain N'ghathrod — that player mills that many cards"
	nghaReanimateLabel = "Captain N'ghathrod — put an artifact or creature card milled this turn onto the battlefield under your control"
)

// Captain N'ghathrod — Legendary Creature — Horror Pirate {3}{U}{B}, 3/6:
//
//	"Horrors you control have menace.
//	 Whenever a Horror you control deals combat damage to a player, that
//	 player mills that many cards.
//	 At the beginning of your end step, choose target artifact or
//	 creature card in an opponent's graveyard that was put there from
//	 their library this turn. Put it onto the battlefield under your
//	 control."
//
// The menace is a tribal layer-6 grant (TribalKeywordGrant) with no
// "other", so the Captain, a Horror, has it too, and it reaches a
// changeling or a creature made a Horror by another effect.
//
// The mill trigger is one per Horror that connects, not "one or more":
// each creature's combat damage event is its own trigger (CR 603.2c),
// and the player it hit mills the damage that creature dealt. Whether
// the creature is a Horror is read off it as the damage is dealt,
// through the Captain's own words (game.ThroughTextOf), as every tribal
// trigger here reads it.
//
// The end-step ability targets (CR 603.3d): an artifact or creature card
// an opponent owns, in their graveyard, whose most recent arrival there
// this turn came out of a library — a mill, a surveil, an Entomb
// (nghaPutThereFromLibraryThisTurn). A card that died, was discarded,
// or was already there when the turn began is not a legal target, and
// with none the trigger is removed. The pick is checked again as it
// resolves (CR 608.2b), and the card enters under the Captain's
// controller's control, whoever owns it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cfb23c6b-6e4a-4fc9-b3bb-a3ddc2ba06b8",
		Name:         "Captain N'ghathrod",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Horror"}, YoursOnly: true}, "menace"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, nghaHorrorDealtCombatDamageToAPlayer, nghaMillLabel, nghaMillThatMany),
			Targeting(
				AtYourEndStep(nghaReanimateLabel, returnFirstLegalGraveyardTargetUnderYourControl),
				TargetCardInGraveyard("target artifact or creature card in an opponent's graveyard that was put there from their library this turn",
					Or(Artifact(), Creature()), Not(YouOwn()), nghaPutThereFromLibraryThisTurn),
			),
		},
	})
}

// nghaHorrorDealtCombatDamageToAPlayer is the mill trigger's condition:
// combat damage to a player from a creature the Captain's controller
// controls that is a Horror as the damage is dealt.
func nghaHorrorDealtCombatDamageToAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !combatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && OfCreatureType("Horror")(g, source.Controller, game.ThroughTextOf(source, c))
}

// nghaMillThatMany is the mill trigger's body: the damaged player mills
// as many cards as the creature dealt, read off the event that fired it.
func nghaMillThatMany(g *game.Game, item *game.StackItem) error {
	ev := item.Trigger.Event
	if ev.Amount <= 0 || g.PlayerByIDForEffect(ev.Target) == nil {
		return nil
	}
	return MillCards{Player: ev.Target, N: ev.Amount}.Apply(NewContext(g, item))
}

// nghaPutThereFromLibraryThisTurn is the end-step target's provenance
// clause: the card's most recent arrival in a graveyard this turn came
// out of a library. discardedOrMilledThisTurn's walk without its discard
// half — every event of one arrival is read together (a mill is an
// EventMill beside the zone move), a later departure from the graveyard
// ends that arrival, and only the last arrival counts. False for a card
// already in the graveyard when the turn began.
//
// Caller holds g.mu.
func nghaPutThereFromLibraryThisTurn(g *game.Game, _ uuid.UUID, c game.Card) bool {
	inArrival, fromLibrary := false, false
	for _, ev := range g.EventsThisTurn() {
		if ev.CardID != c.InstanceID {
			continue
		}
		switch {
		case ev.NewZone == game.ZoneGraveyard && ev.OldZone != game.ZoneGraveyard:
			if !inArrival {
				inArrival, fromLibrary = true, false
			}
			if ev.OldZone == game.ZoneLibrary {
				fromLibrary = true
			}
		case ev.OldZone == game.ZoneGraveyard && ev.NewZone != "" && ev.NewZone != game.ZoneGraveyard:
			inArrival, fromLibrary = false, false
		}
	}
	return inArrival && fromLibrary
}
