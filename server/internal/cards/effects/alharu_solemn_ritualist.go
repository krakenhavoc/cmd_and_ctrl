package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c8deef08-fcf2-41e1-8aeb-aa4e0af147ca",
		Name:         "Alharu, Solemn Ritualist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Alharu, Solemn Ritualist — put a +1/+1 counter on each of up to two other target creatures", SupportEffect),
				SupportTargets(2, true)),
			TriggerWithPurpose(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && !IsToken(dead) && leftUnderControlOf(ev, dead) == source.Controller &&
					g.LastKnownCountersForEffect(ev.CardID)[game.CounterPlusOne] > 0
			}, "Alharu, Solemn Ritualist — create a 1/1 white Spirit creature token with flying",
				Do(CreateToken{Template: TokenCard("1/1 white Spirit with flying"), N: 1})),
				game.Purpose{DeathPayoff: true, Tokens: 1}),
		},
	})
}
