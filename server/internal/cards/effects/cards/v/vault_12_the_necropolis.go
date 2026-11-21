package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
// (putPlusOneCountersOnEachYouControlThat).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc9d75e7-1f64-4376-8e19-f6622ffbeb28",
		Name:         "Vault 12: The Necropolis",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, "Vault 12: The Necropolis — I: each player gets three rad counters",
				func(g *game.Game, item *game.StackItem) error {
					return effects.EachPlayerGetsRadCounters(g, item.Controller, 3)
				}),
			effects.ChapterTrigger(2, "Vault 12: The Necropolis — II: create X 2/2 Zombie Mutant tokens, X the rad counters among players",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{
						Controller: item.Controller,
						Template:   effects.TokenCard("2/2 black Zombie Mutant"),
						N:          effects.RadCountersAmongPlayers(g),
					}.Apply(effects.NewContext(g, item))
				}),
			effects.ChapterTrigger(3, "Vault 12: The Necropolis — III: two +1/+1 counters on each Zombie or Mutant you control",
				effects.PutPlusOneCountersOnEachYouControlThat(2, effects.Or(effects.OfSubtype("Zombie"), effects.OfSubtype("Mutant")))),
		},
	})
}
