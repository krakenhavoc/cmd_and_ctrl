package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scrapper Champion — Creature — Human Artificer {3}{R}, 2/2:
//
//	"Double strike (This creature deals both first-strike and regular combat damage.)
//	 When this creature enters, you get {E}{E} (two energy counters).
//	 Whenever this creature attacks, you may pay {E}{E}. If you do, put a +1/+1 counter on it."
//
// ADR 0129 §3 (#1995): the energy is paid as the attack trigger resolves
// (CR 118.12), through the pay-unless prompt with an energy payment,
// which holds the declare attackers step until it is answered. "It" is
// the creature that attacked (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4fa99dc1-c2d2-4b21-8511-e3d86626609d",
		Name:            "Scrapper Champion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Purpose:         game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Scrapper Champion", 2),
			whenThisAttacksMayPayEnergy("Scrapper Champion", 2, "put a +1/+1 counter on it", thisStillHere(plusOneCountersOnThis(1))),
		},
	})
}
