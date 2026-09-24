package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ichor Rats — Creature — Phyrexian Rat {1}{B}{B}, 2/1:
//
//	"Infect
//	 When this creature enters, each player gets a poison counter."
//
// Infect is the engine's damage tail (ADR 0056, #748). The enter
// trigger gives every player still in the game one poison counter —
// its controller included — through the CR 614 player-counter window,
// placed by the trigger's controller (CR 120.3b's "you give"), so a
// Vorinclex or a Solemnity sees it.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "f8148664-49c0-421f-93a5-cd59e4e9ea36",
		Name:            "Ichor Rats",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"infect"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ichor Rats — each player gets a poison counter", eachPlayerGetsAPoisonCounter),
		},
	})
}

// eachPlayerGetsAPoisonCounter gives every player still in the game
// one poison counter, in seat order, placed by the item's controller.
func eachPlayerGetsAPoisonCounter(g *game.Game, item *game.StackItem) error {
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if err := g.AddPlayerCounterByForEffect(item.Controller, p.ID, game.CounterPoison, 1); err != nil {
			return err
		}
	}
	return nil
}
