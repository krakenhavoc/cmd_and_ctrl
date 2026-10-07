package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Voltaic Brawler — Creature — Human Warrior {R}{G}, 3/2:
//
//	"When this creature enters, you get {E}{E} (two energy counters).
//	 Whenever this creature attacks, you may pay {E}. If you do, it gets +1/+1 and gains trample until end of turn."
//
// ADR 0129 §3 (#1995): the energy is paid as the attack trigger resolves
// (CR 118.12), through the pay-unless prompt with an energy payment,
// which holds the declare attackers step until it is answered. "It" is
// the creature that attacked (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ab7609dd-f4c0-4636-8177-7a926c01e470",
		Name:         "Voltaic Brawler",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Voltaic Brawler", 2),
			whenThisAttacksMayPayEnergy("Voltaic Brawler", 1, "have it get +1/+1 and trample", voltaicBrawlerPump),
		},
	})
}
