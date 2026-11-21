package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Graviton, Fundamental Force — {2}{U} Legendary Creature — Human Villain 3/3
// (Marvel Super Heroes Commander, #4455):
//
//	"Whenever you draw your second card each turn, choose one —
//	 • Target creature gains flying until end of turn.
//	 • Tap target creature."
//
// Kid Loki's second-draw condition on a modal trigger (Tiller Engine's
// shape): the mode and its target are chosen as the trigger goes on the
// stack.
//
// No simplification.
func init() {
	const label = "Graviton, Fundamental Force — choose one: target creature gains flying until end of turn, or tap target creature"
	effects.Register(effects.Spec{
		OracleID:     "eab391d2-ccdd-401c-8f10-69aac4619213",
		Name:         "Graviton, Fundamental Force",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDrawCard},
			Key:     label,
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B41DrewYourSecondCardThisTurn(ev, source, g)
			},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
			Modes: effects.ChooseOne(
				effects.ModeDoing("Target creature gains flying until end of turn.",
					effects.TargetCreature("target creature"), effects.Msc12FlyingTheModesTarget),
				effects.ModeDoing("Tap target creature.",
					effects.TargetCreature("target creature"), effects.TapTheModesTarget),
			),
		}},
	})
}
