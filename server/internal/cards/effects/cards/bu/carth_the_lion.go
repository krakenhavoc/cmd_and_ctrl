package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Carth the Lion — Legendary Creature — Human Warrior {2}{B}{G}, 3/5
// (Modern Horizons 2, #6148):
//
//	"Whenever Carth enters or a planeswalker you control dies, look at the
//	 top seven cards of your library. You may reveal a planeswalker card
//	 from among them and put it into your hand. Put the rest on the bottom
//	 of your library in a random order.
//	 Planeswalkers' loyalty abilities you activate cost an additional [+1]
//	 to activate."
//
// One printed trigger with two conditions, one registered trigger: Carth's
// own entry, or a planeswalker its controller controlled dying, read from
// the planeswalker's last-known types and controller (CR 603.10a). The
// static is Spec.LoyaltyCostAdd: CR 606.5 combines the [+1] with the
// printed loyalty cost into one, so a −2 is a −1, a 0 is a +1, and a −3 is
// activatable at 2 loyalty (CR 606.6 judges the combined cost).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "e0709843-0a91-4a21-9512-d7481707f3c9",
		Name:           "Carth the Lion",
		Completeness:   effects.CompletenessFull,
		LoyaltyCostAdd: 1,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventETB, game.EventLTB}, carthEntersOrAPlaneswalkerDies,
				"Carth the Lion — look at the top seven cards; you may reveal a planeswalker card and put it into your hand, the rest on the bottom in a random order",
				effects.LookAtTopThenMayTakeToHand(7, effects.Planeswalker(), 1,
					"Carth the Lion — you may reveal a planeswalker card and put it into your hand")),
		},
	})
}

// carthEntersOrAPlaneswalkerDies is "whenever Carth enters or a
// planeswalker you control dies".
func carthEntersOrAPlaneswalkerDies(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if ev.Kind == game.EventETB {
		return effects.Self(ev, source, lki, g)
	}
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
		return false
	}
	dead, ok := g.LookupCardForEffect(ev.CardID)
	return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller && effects.LeftAsType(ev, dead, "planeswalker")
}
