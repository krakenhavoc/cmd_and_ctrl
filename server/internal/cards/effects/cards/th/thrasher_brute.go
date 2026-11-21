package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thrasher Brute — Creature — Orc Warrior {3}{B}, 4/3 (Battlebond, #5529):
//
//	"Whenever this creature or another Warrior your team controls enters,
//	 target opponent loses 1 life and you gain 1 life."
//
// "Your team" is "you" (CR 102.4). One trigger, two conditions (Harvestrite
// Host's shape): the Brute's own entry, and any other Warrior permanent
// entering under its controller's control, read through the layers so a
// changeling counts. The drain is Al Bhed Salvagers' body: the target
// opponent loses 1 and the controller gains 1; an opponent who is no longer
// a legal target loses nothing and the ability does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ac09649-3b6c-40fa-81f8-583384c9e0bb",
		Name:         "Thrasher Brute",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID ||
					effects.B10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Warrior"))
			}, "Thrasher Brute — target opponent loses 1 life and you gain 1 life",
				effects.Do(effects.TargetPlayerLosesLife{N: 1}, effects.GainLife{Amount: 1})),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
