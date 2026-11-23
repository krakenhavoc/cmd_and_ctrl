package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Neerdiv, Devious Diver — Legendary Creature — Merfolk Rogue {2}{U}, 2/2
// (Foundations Jumpstart):
//
//	"Whenever Neerdiv becomes tapped, target player mills cards equal to
//	 its power.
//	 Whenever you cast a spell from your graveyard or activate an ability of
//	 a card in your graveyard, draw a card and put a +1/+1 counter on
//	 Neerdiv."
//
// The tap trigger is Annie Flash's "becomes tapped" watcher (a tap effect or
// an attacker that was tapped to attack, ThisBecameTapped). The player is
// targeted as the trigger goes on the stack and the mill is Neerdiv's power
// as the trigger resolves, or the power it last had if it left (CR 608.2h).
// The second line is one ability over two events, cast and activation. The
// cast is Devoted Grafkeeper's (the spell's origin is the graveyard and the
// card is yours); the activation reads the zone the ability was activated
// from, which the announcement stamps on the event (Event.ActivatedFrom), so
// an embalm or eternalize that exiles the card as a cost still counts. The
// draw and the counter are one resolution, in printed order; the counter
// needs Neerdiv on the battlefield, the draw does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b015ec84-95b9-4010-afef-d9d0cacb98dd",
		Name:         "Neerdiv, Devious Diver",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Neerdiv, Devious Diver — target player mills cards equal to its power", effects.S82BnsB51NeerdivMill),
				effects.TargetPlayer("target player")),
			effects.OnAny([]game.EventKind{game.EventCast, game.EventActivateAbility}, effects.S82BnsB51YouUseYourGraveyard,
				"Neerdiv, Devious Diver — draw a card and put a +1/+1 counter on it",
				effects.Do(effects.DrawCards{N: 1}, effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
