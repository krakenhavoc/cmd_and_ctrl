package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Havi, the All-Father — Legendary Creature — God Warrior {3}{R}{G}{W}, 6/6
// (Assassin's Creed, #5948):
//
//	"Havi has indestructible as long as there are four or more historic
//	 cards in your graveyard. (Artifacts, legendaries, and Sagas are
//	 historic.)
//	 Sage Project — Whenever Havi or another legendary creature you control
//	 dies, return target legendary creature card with lesser mana value from
//	 your graveyard to the battlefield tapped."
//
// The keyword is a layer 6 grant to Havi alone, on while the count of historic
// cards (Jhoira's b09IsHistoric: artifact, legendary or Saga) in its
// controller's graveyard is four or more, read on every recompute (CR 611.3a).
//
// The trigger is any legendary creature you controlled dying, Havi included,
// judged as it last existed (CR 603.10a: the last-known supertype and
// controller on the event). "Lesser mana value" is compared with the dead
// creature's mana value as it last existed, carried on the event's snapshot,
// so the clause is built per trigger (TargetsFrom) and chosen as the trigger
// goes on the stack; the dead creature is itself a legal card only if its mana
// value is lower than its own, which it never is. The creature returns
// tapped, under its owner's control (it is in your graveyard).
//
// No simplification.
func init() {
	const label = "Havi, the All-Father — return target legendary creature card with lesser mana value from your graveyard to the battlefield tapped"
	t := On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		dead, ok := diedCreature(ev, g)
		if !ok || leftUnderControlOf(ev, dead) != source.Controller {
			return false
		}
		was, known := ev.WasSupertype("legendary")
		return known && was
	}, label, returnFirstLegalGraveyardTargetToBattlefieldTapped)
	t.TargetsFrom = haviTargets
	Register(Spec{
		OracleID:     "3f58e0f2-7cb5-4a82-b794-071f433fe583",
		Name:         "Havi, the All-Father",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			func() game.StaticAbility {
				s := KeywordGrant(whileYouHave(haviFourHistoric, selfCreature), "indestructible")
				s.Label = "Has indestructible as long as there are four or more historic cards in your graveyard."
				return s
			}(),
		},
		Triggered: []game.TriggeredAbility{t},
	})
}

// haviFourHistoric is "there are four or more historic cards in your
// graveyard" for a player.
func haviFourHistoric(g *game.Game, you uuid.UUID) bool {
	p := g.PlayerByIDForEffect(you)
	if p == nil || p.Graveyard == nil {
		return false
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if b09IsHistoric(c) {
			n++
		}
	}
	return n >= 4
}

// haviTargets is "target legendary creature card with lesser mana value
// from your graveyard": lesser than the creature that died, as it last
// existed.
func haviTargets(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
	mv := 0
	if tc.Object != nil {
		mv = tc.Object.ManaValue
	}
	return TargetCardInGraveyard("target legendary creature card with lesser mana value from your graveyard",
		YouOwn(), Creature(), Legendary(),
		func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.ManaValue() < mv })
}
