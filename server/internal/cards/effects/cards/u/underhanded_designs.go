package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Underhanded Designs — Enchantment {1}{B} (Kaladesh, #5539):
//
//	"Whenever an artifact you control enters, you may pay {1}. If you do, each
//	 opponent loses 1 life and you gain 1 life.
//	 {1}{B}, Sacrifice this enchantment: Destroy target creature. Activate
//	 only if you control two or more artifacts."
//
// The trigger is Mirrorworks' "artifact entered under your control" watcher
// with its "you may pay. If you do" (MayPay), and Blood Artist's drain as the
// paid branch. The activated ability is Yavimaya Steelcrusher's sacrifice-as-
// a-cost destroy with Brotherhood Scribe's count gate: the artifacts are
// counted when it is activated (CR 602.5b), the Designs itself is not one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "870185fc-7b05-47d3-95ff-97fbf12796d4",
		Name:         "Underhanded Designs",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.IsArtifactEnteredUnderYourControl(ev, source, g)
			}, "Underhanded Designs — you may pay {1}; if you do, each opponent loses 1 life and you gain 1 life",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Underhanded Designs — pay {1}: each opponent loses 1 life and you gain 1 life?",
						Then:     underhandedDesignsPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     "{1}{B}, Sacrifice this enchantment: Destroy target creature. Activate only if you control two or more artifacts.",
			Cost:      effects.Plus(effects.ManaCost("{1}{B}"), effects.SacrificeThis()),
			Condition: effects.ControlsAtLeast(2, effects.MatchArtifact),
			Targets:   effects.TargetCreature("target creature"),
			Effect:    effects.DestroyFirstLegalTarget,
		}},
	})
}

// underhandedDesignsPaid is the "if you do": the drain. An on-disk key.
var underhandedDesignsPaid = effects.PayUnlessBranches("pay-unless/underhanded-designs-drain",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		if err := (effects.EachOpponentLosesLife{N: 1}).Apply(ctx); err != nil {
			return err
		}
		return effects.GainLife{Player: ctx.Controller(), Amount: 1}.Apply(ctx)
	}, nil)
