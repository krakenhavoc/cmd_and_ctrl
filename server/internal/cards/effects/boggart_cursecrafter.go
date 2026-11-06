package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boggart Cursecrafter — Creature — Goblin Warlock {B}{R}, 2/3 (Lorwyn
// Eclipsed, #4458):
//
//	"Deathtouch
//	 Whenever another Goblin you control dies, this creature deals 1
//	 damage to each opponent."
//
// Pashalik Mons' "another Goblin you control dies" read
// (b25AnotherGoblinYouControlDied): the creature's controller and types
// are what it had as it died (CR 603.10a), so a changeling counts and so
// does a token that was a Goblin only through a grant; "Goblin" is a word
// of this card's text (CR 612.2). The damage is one event from this
// creature to each opponent (DamageEachOpponent), dealt as it last existed
// if it has left by then. The row declares DeathPayoff, as Cruel
// Celebrant's does (#4990), so the bot keeps it in play beside its
// Goblins.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5d58aefe-a3df-4e0a-b679-b967cd5fd2c6",
		Name:            "Boggart Cursecrafter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b25AnotherGoblinYouControlDied(ev, source, g)
			}, "Boggart Cursecrafter — 1 damage to each opponent", Do(DamageEachOpponent{N: 1})), game.Purpose{DeathPayoff: true}),
		},
	})
}
