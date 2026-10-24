package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vault 12: The Necropolis — Enchantment — Saga {4}{B}{B}:
//
//	"(As this Saga enters and after your draw step, add a lore
//	 counter. Sacrifice after III.)
//	 I — Each player gets three rad counters.
//	 II — Create X 2/2 black Zombie Mutant creature tokens, where X is
//	 the total number of rad counters among players.
//	 III — Put two +1/+1 counters on each creature you control that's
//	 a Zombie or Mutant."
//
// #2809. Chapter I is eachPlayerGetsRadCounters (the Saga's controller
// gives them, themselves included). Chapter II counts the rad counters
// of every player still in the game as the chapter resolves, so the
// counters chapter I just gave are in it, and a rad trigger that milled
// some away between the chapters is not. A player who has left the game
// has no counters (CR 800.4a). Chapter III is one placement of two
// counters on each creature the Saga's controller controls that is a
// Zombie or a Mutant, the tokens from chapter II among them
// (putCountersOnEachYouControlThat).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "bc9d75e7-1f64-4376-8e19-f6622ffbeb28",
		Name:         "Vault 12: The Necropolis",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, "Vault 12: The Necropolis — I: each player gets three rad counters",
				func(g *game.Game, item *game.StackItem) error {
					return eachPlayerGetsRadCounters(g, item.Controller, 3)
				}),
			ChapterTrigger(2, "Vault 12: The Necropolis — II: create X 2/2 Zombie Mutant tokens, X the rad counters among players",
				func(g *game.Game, item *game.StackItem) error {
					return CreateToken{
						Controller: item.Controller,
						Template:   TokenCard("2/2 black Zombie Mutant"),
						N:          radCountersAmongPlayers(g),
					}.Apply(NewContext(g, item))
				}),
			ChapterTrigger(3, "Vault 12: The Necropolis — III: two +1/+1 counters on each Zombie or Mutant you control",
				putCountersOnEachYouControlThat("+1/+1", 2, Or(OfSubtype("Zombie"), OfSubtype("Mutant")))),
		},
	})
}
