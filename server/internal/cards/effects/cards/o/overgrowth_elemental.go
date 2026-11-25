package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Overgrowth Elemental — Creature — Elemental {2}{G}, 3/2 (Core Set 2020,
// #5518):
//
//	"When this creature enters, put a +1/+1 counter on another target Elemental
//	 you control.
//	 Whenever another creature you control dies, you gain 1 life. If that
//	 creature was an Elemental, put a +1/+1 counter on this creature."
//
// The enter trigger's target is chosen as it goes on the stack and re-checked as
// it resolves (CR 608.2b). The dies trigger is one per creature, reading the
// creature's type as it last existed (CR 603.10a), so an Elemental that was
// turned into something else before it died does not count, and one that was
// made an Elemental does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3271e3b9-1710-4dfe-a381-00a3ee1032b2",
		Name:         "Overgrowth Elemental",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Overgrowth Elemental — put a +1/+1 counter on another target Elemental you control", effects.OldB156CountersOnTheTarget(1)),
				effects.Another(effects.TargetCreature("another target Elemental you control", effects.YouControl(), effects.OfSubtype("Elemental")))),
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AnotherCreatureYouControlDied(ev, source, g)
			}, "Overgrowth Elemental — you gain 1 life; if it was an Elemental, put a +1/+1 counter on this creature", effects.OldB156ElementalDied),
		},
	})
}
