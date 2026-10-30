package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "eab391d2-ccdd-401c-8f10-69aac4619213",
		Name:         "Graviton, Fundamental Force",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDrawCard},
			Key:     label,
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
			Modes: ChooseOne(
				ModeDoing("Target creature gains flying until end of turn.",
					TargetCreature("target creature"), msc12FlyingTheModesTarget),
				ModeDoing("Tap target creature.",
					TargetCreature("target creature"), TapTheModesTarget),
			),
		}},
	})
}
