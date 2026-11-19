package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Basri's Lieutenant — Creature — Human Knight {3}{W}, 3/4 (Foundations
// Jumpstart):
//
//	"Vigilance, protection from multicolored
//	 When this creature enters, put a +1/+1 counter on target creature
//	 you control.
//	 Whenever this creature or another creature you control dies, if it
//	 had a +1/+1 counter on it, create a 2/2 white Knight creature token
//	 with vigilance."
//
// "Had a +1/+1 counter on it" reads the dead creature's last-known
// counters (CR 603.10a, 603.4), as Alharu does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "57d0c688-6206-4c83-9fcb-27fd29e2a9ed",
		Name:            "Basri's Lieutenant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "protection from multicolored"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Basri's Lieutenant — put a +1/+1 counter on target creature you control",
					PlusOneCounterOnTheTarget),
				TargetCreature("target creature you control", YouControl())),
			TriggerWithPurpose(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && leftUnderControlOf(ev, dead) == source.Controller &&
					g.LastKnownCountersForEffect(ev.CardID)[game.CounterPlusOne] > 0
			}, "Basri's Lieutenant — create a 2/2 white Knight creature token with vigilance",
				Do(CreateToken{Template: TokenCard("2/2 white Knight with vigilance"), N: 1})),
				game.Purpose{DeathPayoff: true, Tokens: 1}),
		},
	})
}
