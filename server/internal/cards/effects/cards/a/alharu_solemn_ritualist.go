package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alharu, Solemn Ritualist — Legendary Creature — Human Monk {4}{W}, 3/3:
//
//	"When Alharu enters, put a +1/+1 counter on each of up to two other
//	 target creatures.
//	 Whenever a nontoken creature you control with a +1/+1 counter on it
//	 dies, create a 1/1 white Spirit creature token with flying.
//	 Partner (You can have two commanders if both have partner.)"
//
// The enter trigger is the support shape (SupportTargets / SupportEffect).
// The dies trigger reads the counters off the creature as it last existed
// (CR 603.10a), so a creature that died with a counter counts, Alharu
// included, and a token never does. Partner is a deck-building rule
// (CR 702.124) that internal/deck reads off the oracle text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8deef08-fcf2-41e1-8aeb-aa4e0af147ca",
		Name:         "Alharu, Solemn Ritualist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Alharu, Solemn Ritualist — put a +1/+1 counter on each of up to two other target creatures", effects.SupportEffect),
				effects.SupportTargets(2, true)),
			effects.TriggerWithPurpose(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && !effects.IsToken(dead) && effects.LeftUnderControlOf(ev, dead) == source.Controller &&
					g.LastKnownCountersForEffect(ev.CardID)[game.CounterPlusOne] > 0
			}, "Alharu, Solemn Ritualist — create a 1/1 white Spirit creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Spirit with flying"), N: 1})),
				game.Purpose{DeathPayoff: true, Tokens: 1}),
		},
	})
}
