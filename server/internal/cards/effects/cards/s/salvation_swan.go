package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Salvation Swan — Creature — Bird Soldier {3}{W}, 3/3 (Bloomburrow, #4467):
//
//	"Flash
//	 Flying
//	 Whenever this creature or another Bird you control enters, exile up to
//	 one target creature you control without flying. Return it to the
//	 battlefield under its owner's control with a flying counter on it at
//	 the beginning of the next end step."
//
// One trigger on every Bird entering under your control, itself included.
// The exile is one event with the delayed CR 603.7 return; the card comes
// back a new object with the flying counter put on as it enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0f68ae67-1671-4281-8b6b-3e52fdd4915f",
		Name:            "Salvation Swan",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.HasSubtype(source.TextSubtype("Bird"))
			}, "Salvation Swan — exile up to one target creature you control without flying; return it with a flying counter at the beginning of the next end step",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ExileTargetsThenScheduleReturnBody(effects.NewContext(g, item),
						"Salvation Swan — return the exiled creature with a flying counter",
						effects.ReturnExiledWithFlyingCounterBody)
				}),
				effects.TargetCreature("up to one target creature you control without flying", effects.YouControl(), effects.WithoutKeyword("flying")).WithCount(0, 1)),
		},
	})
}
