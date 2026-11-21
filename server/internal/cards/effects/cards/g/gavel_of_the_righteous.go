package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gavel of the Righteous — Artifact — Equipment {2} (Foundations
// Commander):
//
//	"At the beginning of combat on your turn, put a charge counter on this
//	 Equipment.
//	 Equipped creature gets +1/+1 for each counter on this Equipment.
//	 As long as this Equipment has four or more counters on it, equipped
//	 creature has double strike.
//	 Equip—Pay {3} or remove a counter from this Equipment."
//
// Additive Evolution's beginning-of-combat watcher puts the counter. The
// bonus counts every counter on the Equipment (Malefic Scythe's
// PumpAttachedPer, read live), and double strike is a keyword grant gated
// on that count (GrantToAttachedWhile). Equip is an activated ability whose
// cost is an either/or (OrCost, #4566): {3}, or removing a counter of any
// kind from the Equipment itself, announced with the activation.
//
// No simplification.
func init() {
	totalCounters := func(_ *game.Game, source *game.Card) int {
		n := 0
		for _, c := range source.Counters {
			n += c
		}
		return n
	}
	effects.Register(effects.Spec{
		OracleID:     "204a5616-9f19-4d4c-aeb0-d1ec486b9ace",
		Name:         "Gavel of the Righteous",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, totalCounters),
			effects.GrantToAttachedWhile(func(_ *game.Card, g *game.Game, source *game.Card) bool {
				return totalCounters(g, source) >= 4
			}, "double strike"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventStepBegan, effects.StepBegan(game.StepBeginCombat, true),
				"Gavel of the Righteous — put a charge counter on this Equipment",
				effects.B33PutChargeCounterOnSelf),
		},
		Activated: []effects.ActivatedAbility{effects.EquipPayingAbility("Equip—Pay {3} or remove a counter from this Equipment",
			effects.OrCost(effects.ManaBranch("{3}"), effects.RemoveCountersBranch("", 1, "a counter from this Equipment")))},
	})
}
