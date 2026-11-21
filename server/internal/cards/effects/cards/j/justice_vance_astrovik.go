package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Justice, Vance Astrovik — Legendary Creature — Mutant Hero {2}{U}, 2/2
// (Marvel Super Heroes, #4455):
//
//	"Flying
//	 When Justice enters, return up to one target nonland, nontoken
//	 permanent to its owner's hand.
//	 Whenever another nonland permanent you control is returned to its
//	 owner's hand, put a +1/+1 counter on Justice."
//
// The enters trigger is Skyclave Apparition's target clause (up to one, any
// controller) with Sanctum of Eternity's bounce body, which goes through the
// shared exit primitive. The counter trigger watches a permanent leaving the
// battlefield for a hand (EventLTB with NewZone hand): the permanent must be
// another one than Justice, nonland, and under the controller's control as it
// left (CR 603.10a; the event's last-known information). A token that is
// bounced does move to the hand for an instant, so it counts. Justice's own
// bounce is excluded by "another".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "184dba74-63ed-4820-9174-68adf26bfabf",
		Name:            "Justice, Vance Astrovik",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Justice — return up to one target nonland, nontoken permanent to its owner's hand", effects.BounceChosenTarget),
				effects.TargetPermanent("up to one target nonland, nontoken permanent", effects.Nonland(), effects.Not(effects.IsTokenPredicate())).WithCount(0, 1)),
			effects.On(game.EventLTB, justiceAnotherNonlandReturned,
				"Justice — a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}

// justiceAnotherNonlandReturned is "another nonland permanent you control is
// returned to its owner's hand": it left the battlefield for a hand, was not
// Justice, was not a land as it last existed, and was under the source
// controller's control as it left.
func justiceAnotherNonlandReturned(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneHand || ev.CardID == source.InstanceID {
		return false
	}
	c, found := g.LookupCardForEffect(ev.CardID)
	controller, known := ev.LeftUnderControlOf()
	if !known {
		if !found {
			return false
		}
		controller = c.Controller
	}
	if controller != source.Controller {
		return false
	}
	if was, known := ev.WasType("land"); known {
		return !was
	}
	return found && !c.HasCardType("land")
}
