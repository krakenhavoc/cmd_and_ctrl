package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disa the Restless — Legendary Creature — Human Scout {2}{B}{R}{G}, 5/6:
//
//	"Whenever a Lhurgoyf permanent card is put into your graveyard from
//	 anywhere other than the battlefield, put it onto the battlefield.
//	 Whenever one or more creatures you control deal combat damage to a
//	 player, create a Tarmogoyf token."
//
// The first trigger is Syr Konrad's graveyard-arrival test: a discard, a
// mill, or a spell leaving the stack for the controller's graveyard (a
// death is excluded by the origin), for a permanent card with the Lhurgoyf
// type read through the source's text (CR 612.2). It puts the card onto the
// battlefield from the graveyard if it is still there as the trigger
// resolves. The second is the "one or more" batch trigger (CR 603.2c) of
// Whenever One Or More Creatures You Control Deal Combat Damage to a
// Player, making Disa's Tarmogoyf token (TarmogoyfToken).
func init() {
	const name = "Disa the Restless"
	effects.Register(effects.Spec{
		OracleID:     "283bfc53-afd3-4d21-b7d5-5baba4232422",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventDiscardCard, game.EventMill, game.EventZoneMove},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.NewZone != game.ZoneGraveyard || ev.OldZone == game.ZoneBattlefield || ev.OldZone == game.ZoneGraveyard {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.Owner == source.Controller && c.IsPermanent() && c.HasSubtype(source.TextSubtype("Lhurgoyf"))
				}, name+" — put the Lhurgoyf permanent card onto the battlefield",
				effects.S82OldB062PutThatCardOntoTheBattlefield),
			effects.WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(effects.Creature(),
				name+" — create a Tarmogoyf token",
				effects.Do(effects.CreateToken{Template: effects.TarmogoyfToken(), N: 1})),
		},
	})
}
