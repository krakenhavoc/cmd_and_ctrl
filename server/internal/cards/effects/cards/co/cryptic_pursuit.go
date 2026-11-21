package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cryptic Pursuit — Enchantment {2}{U}{R} (Streets of New Capenna Commander):
//
//	"Whenever you cast an instant or sorcery spell from your hand, manifest
//	 the top card of your library. (Put that card onto the battlefield face
//	 down as a 2/2 creature. Turn it face up any time for its mana cost if
//	 it's a creature card.)
//	 Whenever a face-down creature you control dies, exile it if it's an
//	 instant or sorcery card. You may cast that card until the end of your
//	 next turn."
//
// The first trigger is Invocation of the Founders' read of the cast's zone
// (a flashback or exile cast never triggers it) with Primordial Mist's
// manifest. The second is Yarus's "face-down creature you control dies"
// (the last-known face-down stamp, CR 708.9, since the card in the graveyard
// is face up). As it resolves, the card must still be in a graveyard and be
// an instant or sorcery card; it is exiled and its controller may cast it
// for its mana cost until the end of their next turn (Bruse Tarl's seat-turn
// window, ADR 0063). A creature or land card stays in the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "214cfbc8-369e-4035-9b65-80d0e0024889",
		Name:         "Cryptic Pursuit",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller || ev.OldZone != game.ZoneHand {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && (c.IsInstant() || c.IsSorcery())
			}, "Cryptic Pursuit — manifest the top card of your library", effects.Do(effects.Manifest{N: 1})),
			effects.On(game.EventLTB, effects.YarusFaceDownCreatureDied,
				"Cryptic Pursuit — exile it if it's an instant or sorcery card; you may cast it until the end of your next turn",
				crypticPursuitExile),
		},
	})
}

func crypticPursuitExile(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	z := g.FindCardZoneForEffect(id)
	c, ok := g.LookupCardForEffect(id)
	if z == nil || z.Kind != game.ZoneGraveyard || !ok || !(c.IsInstant() || c.IsSorcery()) {
		return nil
	}
	return g.ExileCardWithPermissionForEffect(id, game.CastPermission{
		Player:   item.Controller,
		CastOnly: true,
		Duration: g.UntilEndOfYourNextTurnDuration(item.Controller),
	})
}
